package mongo

import (
	"fmt"
	"regexp"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/bondhan/golib/constant"

	"github.com/bondhan/golib/docstore"
)

func toMongoFilter(q *docstore.QueryOpt) (interface{}, *options.FindOptions) {
	d := bson.D{}

	if q == nil {
		return d, nil
	}

	for _, s := range q.Filter {
		if s.Ops == constant.OR {
			if fArr, ok := s.Value.([]docstore.FilterOpt); ok {
				a := toMongoA(fArr)
				d = append(d, bson.E{Key: "$or", Value: a})
			}
			continue
		}

		d = append(d, bson.E{Key: s.Field, Value: toMongoD(s)})
	}

	if q.Page > 0 && q.Limit > 0 {
		q.Skip = q.Page * q.Limit
	}

	opt := options.Find()
	opt.SetLimit(int64(q.Limit))
	opt.SetSkip(int64(q.Skip))
	if q.OrderBy != "" {
		dir := -1
		if q.IsAscend {
			dir = 1
		}
		opt.SetSort(bson.D{{Key: q.OrderBy, Value: dir}})
	}

	return d, opt
}

func toMongoA(f []docstore.FilterOpt) bson.A {
	a := bson.A{}

	for _, v := range f {
		d := bson.D{}
		if v.Ops == constant.AND {
			af, ok := v.Value.([]docstore.FilterOpt)
			if !ok {
				continue
			}
			for i := range af {
				d = append(d, primitive.E{af[i].Field, toMongoD(af[i])})
			}
		} else {
			d = bson.D{
				{Key: v.Field, Value: toMongoD(v)},
			}
		}

		a = append(a, d)
	}

	return a
}

func toMongoD(f docstore.FilterOpt) bson.D {
	switch f.Ops {
	case constant.EX:
		return bson.D{
			{Key: "$exists", Value: f.Value},
		}
	case constant.EQ, constant.SE:
		return bson.D{
			{Key: "$eq", Value: f.Value},
		}
	case constant.LT:
		return bson.D{
			{Key: "$lt", Value: f.Value},
		}
	case constant.LE:
		return bson.D{
			{Key: "$lte", Value: f.Value},
		}
	case constant.GT:
		return bson.D{
			{Key: "$gt", Value: f.Value},
		}
	case constant.GE:
		return bson.D{
			{Key: "$gte", Value: f.Value},
		}
	case constant.NE:
		return bson.D{
			{Key: "$ne", Value: f.Value},
		}
	case constant.IN, constant.AIN, constant.AM:
		return bson.D{
			{Key: "$in", Value: f.Value},
		}
	case constant.NIN:
		return bson.D{
			{Key: "$nin", Value: f.Value},
		}
	case constant.EM:
		return bson.D{{
			Key:   "$elemMatch",
			Value: toElemFilter(f.Value),
		}}
	case constant.RE:
		return bson.D{{
			Key: "$regex",
			Value: primitive.Regex{
				Pattern: regexp.QuoteMeta(fmt.Sprintf("%v", f.Value)),
				Options: "i",
			}},
		}
	default:
		return nil
	}
}

func toMongoE(fields docstore.Field) bson.D {
	var d bson.D
	_, ok := fields.Value.(docstore.Field)
	if !ok {
		d = append(d, bson.E{Key: fields.Name, Value: fields.Value})
		return d
	}

	d = append(d, bson.E{Key: fields.Name, Value: toMongoE(fields.Value.(docstore.Field))})
	return d
}

func toElemFilter(val interface{}) bson.D {
	out := bson.D{}
	switch v := val.(type) {
	case []docstore.FilterOpt:
		for _, f := range v {
			out = append(out, bson.E{Key: f.Field, Value: toMongoD(f)})
		}
	}
	return out
}
