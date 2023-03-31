package mongo

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/bondhan/golib/client"
	"github.com/bondhan/golib/config"
	"github.com/bondhan/golib/constant"
	"github.com/bondhan/golib/docstore"
	"github.com/bondhan/golib/util"

	"github.com/bondhan/golib/index"
)

func init() {
	index.Register("atlas", NewMongoAtlas)
}

type AtlasSearch struct {
	store      *mongo.Collection
	Collection string             `json:"collection"`
	Database   string             `json:"database"`
	Indexes    map[string]string  `json:"indexes"`
	Connection client.MongoClient `json:"connection"`
	IndexName  string             `json:"index_name"`
}

type totalCount struct {
	Total int `json:"total"`
}

func NewMongoAtlas(conf interface{}) (index.Driver, error) {
	return NewAtlasSearch(conf)
}

func NewAtlasSearch(conf interface{}) (*AtlasSearch, error) {
	var as AtlasSearch
	switch c := conf.(type) {
	case map[string]interface{}:
		if err := util.DecodeJSON(c, &as); err != nil {
			return nil, err
		}
	case config.Getter:
		if err := c.Unmarshal(&as); err != nil {
			return nil, err
		}
	case AtlasSearch:
		as = c
	case *AtlasSearch:
		as = *c
	default:
		return nil, errors.New("[index/mongo] unsupported connection type")
	}

	client, err := as.Connection.MongoConnect()
	if err != nil {
		return nil, err
	}
	db := client.Database(as.Database)
	as.store = db.Collection(as.Collection)
	return &as, nil
}

func (s *AtlasSearch) Search(ctx context.Context, text string, query *docstore.QueryOpt) (index.SearchResult, error) {
	opts := options.Aggregate().SetMaxTime(5 * time.Second)
	stages := s.buildAggregate(text, query, false)
	res := &AtlasResults{
		scorePath: "score",
		count:     -1,
	}
	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer wg.Done()
		total, err := s.Count(ctx, text, query)
		if err == nil {
			res.count = total
		}
	}()

	cursor, err := s.store.Aggregate(ctx, *stages, opts)
	if err != nil {
		return nil, err
	}

	res.cursor = cursor

	wg.Wait()

	return res, nil
}

func (s *AtlasSearch) Count(ctx context.Context, text string, query *docstore.QueryOpt) (int, error) {
	opts := options.Aggregate().SetMaxTime(5 * time.Second)
	stages := s.buildAggregate(text, query, true)
	cursor, err := s.store.Aggregate(ctx, *stages, opts)
	if err != nil {
		return 0, err
	}

	var count totalCount
	if cursor.Next(ctx) {
		if err := cursor.Decode(&count); err != nil {
			return 0, err
		}
	}
	return count.Total, nil
}

func (s *AtlasSearch) buildAggregate(query string, opt *docstore.QueryOpt, count bool) *mongo.Pipeline {

	shoulds := []bson.D{}

	for k, v := range s.Indexes {
		if v == "fuzzy" {
			shoulds = append(shoulds, bson.D{bson.E{Key: "text", Value: bson.M{
				"path":  k,
				"query": query,
				"fuzzy": bson.M{},
			}}})
			continue
		}
		s := bson.D{bson.E{Key: v, Value: bson.M{
			"path":  k,
			"query": query,
		}}}
		shoulds = append(shoulds, s)
	}

	stages := mongo.Pipeline{
		bson.D{{Key: "$search", Value: bson.M{
			"index": s.IndexName,
			"compound": bson.M{
				"should":             shoulds,
				"minimumShouldMatch": 1,
			}},
		}},
		bson.D{{Key: "$addFields", Value: bson.M{
			"score": bson.M{"$meta": "searchScore"},
			// "count": "$$SEARCH_META.count.total",
		}}},
	}

	if opt != nil {

		if len(opt.Filter) > 0 {
			d := bson.M{}
			for _, s := range opt.Filter {
				d[s.Field] = toMongoM(s)
			}

			stages = append(stages, bson.D{{Key: "$match", Value: d}})
		}

		if !count {
			if opt.Limit > 0 && opt.Page > 0 {
				opt.Skip = opt.Limit * opt.Page
			}

			if opt.Skip > 0 {
				stages = append(stages, bson.D{{Key: "$skip", Value: opt.Skip}})
			}
			if opt.Limit > 0 {
				stages = append(stages, bson.D{{Key: "$limit", Value: opt.Limit}})
			}

			if opt.OrderBy != "" {
				order := -1
				if opt.IsAscend {
					order = 1
				}
				stages = append(stages, bson.D{{Key: "$sort", Value: bson.M{opt.OrderBy: order}}})
			}
		}

	}

	if count {
		stages = append(stages, bson.D{{Key: "$count", Value: "total"}})
	}

	return &stages

}

func toMongoM(f docstore.FilterOpt) bson.M {
	switch f.Ops {
	case constant.EQ, constant.SE:
		return bson.M{
			"$eq": f.Value,
		}
	case constant.LT:
		return bson.M{
			"$lt": f.Value,
		}
	case constant.LE:
		return bson.M{
			"$lte": f.Value,
		}
	case constant.GT:
		return bson.M{
			"$gt": f.Value,
		}
	case constant.GE:
		return bson.M{
			"$gte": f.Value,
		}
	case constant.NE:
		return bson.M{
			"$ne": f.Value,
		}
	case constant.IN, constant.AIN, constant.AM:
		return bson.M{
			"$in": f.Value,
		}
	case constant.EM:
		return bson.M{
			"$elemMatch": f.Value,
		}
	case constant.RE:
		return bson.M{
			"$regex": primitive.Regex{
				Pattern: regexp.QuoteMeta(fmt.Sprintf("%v", f.Value)),
				Options: "i",
			},
		}
	default:
		return nil
	}
}
