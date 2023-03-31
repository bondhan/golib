package mongo

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"

	"github.com/bondhan/golib/client"
	"github.com/bondhan/golib/log"
	"github.com/bondhan/golib/util"

	"github.com/bondhan/golib/docstore"
)

const (
	defaultTTL                    = 24 * time.Hour
	IndexTypeSortOrderAscending   = 1
	IndexTypeSortOrderDescending  = -1
	IndexTypeText                 = "text"
	IndexOptionName               = "name"
	IndexOptionUnique             = "unique"
	IndexOptionExpireAfterSeconds = "expireAfterSeconds"
	IndexOptionSparse             = "sparse"
	defaultIDIndexName            = "_id_"
)

type MongoStore struct {
	store      *mongo.Collection
	idField    string
	collection string
}

func init() {
	docstore.RegisterDriver("mongo", MongoStoreFactory)
}

func MongoStoreFactory(config *docstore.Config) (docstore.Driver, error) {
	return NewMongoStore(config)
}

func NewMongoStore(config *docstore.Config) (*MongoStore, error) {
	switch con := config.Connection.(type) {
	case *client.MongoClient:
		client, err := con.MongoConnect()
		if err != nil {
			return nil, err
		}
		db := client.Database(config.Database)
		return NewMongostore(db, config.Collection, config.IDField)
	case *mongo.Client:
		db := con.Database(config.Database)
		return NewMongostore(db, config.Collection, config.IDField)
	case *mongo.Database:
		return NewMongostore(con, config.Collection, config.IDField)
	case map[string]interface{}:
		var mc client.MongoClient
		if err := util.DecodeJSON(config.Connection, &mc); err != nil {
			return nil, err
		}
		client, err := mc.MongoConnect()
		if err != nil {
			return nil, err
		}
		db := client.Database(config.Database)
		return NewMongostore(db, config.Collection, config.IDField)
	default:
		return nil, errors.New("[docstore/mongo] unsupported connection type")
	}
}

func NewMongostore(db *mongo.Database, collection, idField string) (*MongoStore, error) {
	return &MongoStore{
		store:      db.Collection(collection),
		idField:    idField,
		collection: collection,
	}, nil
}

func (m *MongoStore) Client() *mongo.Client {
	return m.store.Database().Client()
}

func (m *MongoStore) getID(doc interface{}) (interface{}, error) {
	idf := m.idField
	if util.IsStructOrPointerOf(doc) {
		var err error
		idf, err = util.FindFieldByTag(doc, "json", m.idField)
		if err != nil {
			return nil, err
		}
	}

	id, ok := util.Lookup(idf, doc)

	if !ok {
		return nil, errors.New("[docstore/mongo] missing document ID")
	}

	return id, nil
}

func (m *MongoStore) Create(ctx context.Context, doc interface{}) error {
	id, err := m.getID(doc)
	if err != nil {
		return err
	}

	if m.exist(ctx, id) {
		return errors.New("[docstore/mongo] document already exist")
	}

	d := make(map[string]interface{})
	if err := util.DecodeJSON(doc, &d); err != nil {
		return err
	}
	convertTime(d)

	_, err = m.store.InsertOne(ctx, d)
	if err != nil {
		return err
	}

	return nil
}

func (m *MongoStore) Update(ctx context.Context, id, doc interface{}, replace bool) error {

	if replace {
		_, err := m.store.ReplaceOne(ctx, bson.D{{Key: m.idField, Value: id}}, doc)
		return err
	}
	return m.update(ctx, id, doc, false)
}

func (m *MongoStore) Upsert(ctx context.Context, id, doc interface{}) error {
	return m.update(ctx, id, doc, true)
}

func (m *MongoStore) update(ctx context.Context, id, doc interface{}, upsert bool) error {
	out := make(map[string]interface{})
	if err := util.DecodeJSON(doc, &out); err != nil {
		return err
	}

	fields := bson.D{}
	for k, v := range out {
		fields = append(fields, bson.E{Key: k, Value: v})
	}

	update := bson.D{{Key: "$set", Value: fields}}
	opts := options.Update().SetUpsert(upsert)

	res, err := m.store.UpdateOne(ctx, bson.D{{Key: m.idField, Value: id}}, update, opts)
	if err != nil {
		return err
	}

	if res.UpsertedCount == 0 && res.ModifiedCount == 0 {
		if res.MatchedCount == 0 {
			return docstore.NotFound
		}
		return docstore.NothingUpdated
	}

	return nil
}

func (m *MongoStore) UpdateMany(ctx context.Context, filters []docstore.FilterOpt, fields map[string]interface{}) error {
	fld, flt := bson.D{}, bson.D{}

	for k, v := range fields {
		fld = append(fld, bson.E{Key: k, Value: v})
	}

	for _, filter := range filters {
		flt = append(flt,
			bson.E{
				Key: filter.Field,
				Value: bson.D{
					{
						Key:   filter.Ops,
						Value: filter.Value,
					},
				},
			})
	}

	u := bson.D{{Key: "$set", Value: fld}}

	res, err := m.store.UpdateMany(ctx, flt, u)
	if err != nil {
		return err
	}

	if res.UpsertedCount == 0 && res.ModifiedCount == 0 {
		if res.MatchedCount == 0 {
			return docstore.NotFound
		}
		return docstore.NothingUpdated
	}

	return nil
}

func (m *MongoStore) UpdateField(ctx context.Context, id interface{}, fields []docstore.Field) error {
	fs := bson.D{}
	for _, v := range fields {
		fs = append(fs, bson.E{Key: v.Name, Value: v.Value})
	}

	update := bson.D{{Key: "$set", Value: fs}}

	res, err := m.store.UpdateOne(ctx, bson.D{{Key: m.idField, Value: id}}, update)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return docstore.NotFound
	}
	return err
}

func (m *MongoStore) Increment(ctx context.Context, id interface{}, key string, value int) error {
	update := bson.D{{Key: "$inc", Value: bson.D{{Key: key, Value: value}}}}
	upsert := true
	res, err := m.store.UpdateOne(ctx, bson.D{{Key: m.idField, Value: id}}, update, &options.UpdateOptions{Upsert: &upsert})
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return docstore.NotFound
	}
	return err
}

func (m *MongoStore) GetIncrement(ctx context.Context, id interface{}, key string, value int, doc interface{}) error {
	update := bson.D{{Key: "$inc", Value: bson.D{{Key: key, Value: value}}}}
	rp := options.After
	upsert := true
	res := m.store.FindOneAndUpdate(ctx, bson.D{{Key: m.idField, Value: id}}, update, &options.FindOneAndUpdateOptions{ReturnDocument: &rp, Upsert: &upsert})
	return res.Decode(doc)
}

func (m *MongoStore) Delete(ctx context.Context, id interface{}) error {
	_, err := m.store.DeleteOne(ctx, bson.D{{Key: m.idField, Value: id}})
	return err
}

func (m *MongoStore) DeleteMany(ctx context.Context, query *docstore.QueryOpt) error {
	f, _ := toMongoFilter(query)

	res, err := m.store.DeleteMany(ctx, f)
	if err != nil {
		return err
	}

	if res.DeletedCount == 0 {
		return docstore.NotFound
	}

	return err
}

func (m *MongoStore) Get(ctx context.Context, id interface{}, doc interface{}) error {
	out := make(map[string]interface{})
	if err := m.store.FindOne(ctx, bson.D{{Key: m.idField, Value: id}}).Decode(&out); err != nil {
		if err == mongo.ErrNoDocuments {
			return docstore.NotFound
		}
		return err
	}
	out[m.idField] = id

	return util.DecodeJSON(out, doc)
}

func (m *MongoStore) exist(ctx context.Context, id interface{}) bool {
	out := make(map[string]interface{})
	if err := m.store.FindOne(ctx, bson.D{{Key: m.idField, Value: id}}).Decode(&out); err != nil {
		return false
	}
	return true
}

func (m *MongoStore) Count(ctx context.Context, query *docstore.QueryOpt) (int64, error) {
	f, _ := toMongoFilter(query)
	return m.store.CountDocuments(ctx, f)
}

func (m *MongoStore) Find(ctx context.Context, query *docstore.QueryOpt, docs interface{}) error {

	f, opt := toMongoFilter(query)

	res, err := m.store.Find(ctx, f, opt)
	if err != nil {
		return err
	}

	var out []map[string]interface{}

	if err := res.All(ctx, &out); err != nil {
		return err
	}

	return util.DecodeJSON(out, docs)
}

func (m *MongoStore) FindOne(ctx context.Context, query *docstore.QueryOpt, doc interface{}) error {
	f, _ := toMongoFilter(query)
	out := make(map[string]interface{})
	if err := m.store.FindOne(ctx, f).Decode(&out); err != nil {
		if err == mongo.ErrNoDocuments {
			return docstore.NotFound
		}
		return err
	}

	return util.DecodeJSON(out, doc)
}

func (m *MongoStore) Query(ctx context.Context, query *docstore.QueryOpt) (docstore.Iterator, error) {
	f, opt := toMongoFilter(query)
	res, err := m.store.Find(ctx, f, opt)
	if err != nil {
		return nil, err
	}
	return NewMongoIterator(res), nil
}

func (m *MongoStore) BulkCreate(ctx context.Context, docs []interface{}, opts ...interface{}) error {
	ins := make([]interface{}, 0)
	for _, doc := range docs {
		d := make(map[string]interface{})
		if err := util.DecodeJSON(doc, d); err != nil {
			return err
		}
		convertTime(d)
		ins = append(ins, d)
	}

	optsx := make([]*options.InsertManyOptions, 0)

	for _, opt := range opts {
		optInsert := options.InsertMany()
		if opt != nil {
			val, ok := opt.(*docstore.InsertManyOpt)
			if ok {
				if val != nil {
					if val.Ordered != nil {
						optInsert.SetOrdered(*val.Ordered)
					}

					if val.BypassDocumentValidation != nil {
						optInsert.SetBypassDocumentValidation(*val.BypassDocumentValidation)
					}
					optInsert.SetComment(val.Comment)
				}
			}
		}
		optsx = append(optsx, optInsert)
	}

	_, err := m.store.InsertMany(ctx, ins, optsx...)
	return err
}

func (m *MongoStore) BulkGet(ctx context.Context, ids []interface{}, docs interface{}) error {

	res, err := m.store.Find(ctx, bson.D{{Key: m.idField, Value: bson.M{"$in": ids}}})
	if err != nil {
		return err
	}

	var out []map[string]interface{}

	if err := res.All(ctx, &out); err != nil {
		return err
	}

	return util.DecodeJSON(out, docs)
}

func (m *MongoStore) Migrate(ctx context.Context, config interface{}) error {
	logger := log.GetLogger(ctx, "docstore/mongo", "Migrate")
	db := m.store.Database()
	cols, err := db.ListCollectionNames(ctx, bson.D{})
	if err != nil {
		return err
	}
	create := true
	for _, c := range cols {
		if c == m.collection {
			create = false
			break
		}
	}
	if create {
		logger.WithField("collection", m.collection).Info("creating collection")
		if err := db.CreateCollection(ctx, m.collection); err != nil {
			logger.WithError(err).Error("error creating collection")
			return err
		}

	}

	if config == nil {
		return nil
	}

	var dbConf docstore.Config

	switch v := config.(type) {
	case *docstore.Config:
		dbConf = *v
	case docstore.Config:
		dbConf = v
	default:
		logger.Error("not docstore config")
		return nil
	}

	idxs := dbConf.Indexes
	if idxs == nil {
		idxs = make([]map[string]map[string]interface{}, 0)
	}

	if dbConf.IDField != "_id" && !isFieldExist(dbConf.IDField, dbConf.Indexes) {
		idxs = append(idxs, map[string]map[string]interface{}{
			"keys": {
				dbConf.IDField: IndexTypeSortOrderAscending,
			},
			"options": {
				IndexOptionUnique: true,
			},
		})
	}

	if dbConf.TimestampField != "" && !isFieldExist(dbConf.TimestampField, dbConf.Indexes) {
		idxs = append(idxs, map[string]map[string]interface{}{
			"keys": {
				dbConf.TimestampField: IndexTypeSortOrderDescending,
			},
		})
	}

	if dbConf.UpdateTimeField != "" && !isFieldExist(dbConf.UpdateTimeField, dbConf.Indexes) {
		idxs = append(idxs, map[string]map[string]interface{}{
			"keys": {
				dbConf.UpdateTimeField: IndexTypeSortOrderDescending,
			},
		})
	}

	if len(idxs) > 0 {
		specs, err := m.listIndex(ctx)
		if err != nil {
			logger.WithError(err).Error("error listing existing index")
			return nil
		}

		models := make([]mongo.IndexModel, 0)
		mods := make(map[string]struct{})
		confIndexes := make(map[string]struct{})
		for _, idx := range idxs {
			newSpec := m.toIndexSpecification(idx)
			confIndexes[newSpec.Name] = struct{}{}
			if isIndexSpecExist(newSpec, specs) {
				continue
			}

			if isNew(newSpec, specs) {
				if dbConf.DropExistingIndex {
					_, err := m.store.Indexes().DropOne(ctx, newSpec.Name)
					if err != nil {
						return err
					}
				}
			}

			if _, ok := mods[newSpec.Name]; ok {
				continue
			}

			models = append(models, toIndexModel(idx))
			mods[newSpec.Name] = struct{}{}
		}

		if dbConf.DropExistingIndex {
			for _, spec := range specs {
				if spec.Name != defaultIDIndexName {
					if _, ok := confIndexes[spec.Name]; !ok {
						_, err := m.store.Indexes().DropOne(ctx, spec.Name)
						if err != nil {
							return err
						}
					}
				}
			}
		}

		if len(models) > 0 {
			res, err := m.store.Indexes().CreateMany(ctx, models)
			if err != nil {
				logger.WithError(err).Error("error creating index")
				return err
			}
			logger.WithField("created_indexes", res).Info("successfully creating index")
		}
	}

	return nil
}

func isFieldExist(field string, idxs []map[string]map[string]interface{}) bool {
	for _, idx := range idxs {
		for k, v := range idx {
			if k == "keys" {
				if _, ok := v[field]; ok {
					return true
				}
			}
		}
	}
	return false
}

func (m *MongoStore) As(i interface{}) bool {
	p, ok := i.(**mongo.Collection)
	if !ok {
		return false
	}
	*p = m.store
	return true
}

func (m *MongoStore) listIndex(ctx context.Context) ([]*mongo.IndexSpecification, error) {
	return m.store.Indexes().ListSpecifications(ctx)
}

func (m *MongoStore) Ping(ctx context.Context) error {
	return m.store.Database().Client().Ping(ctx, readpref.PrimaryPreferred())
}

func (m *MongoStore) Distinct(ctx context.Context, fieldName string, filter interface{}) ([]interface{}, error) {
	return m.store.Distinct(ctx, fieldName, filter)
}

func (m *MongoStore) Disconnect(ctx context.Context) error {
	client := m.store.Database().Client()
	if client != nil {
		return m.store.Database().Client().Disconnect(ctx)
	}

	return errors.New("client is nil")
}

/*
Pull remove a top level element from an array either by value or object value

	condition = docstore.Field{
		Name:  "menus.id",
		Value: "id to filter document",
	}

	removeCondition = docstore.Field{
		Name: "menus",
		Value: docstore.Field{
			Name:  "id",
			Value: "id to remove array element",
		},
	}

Docs

	https://www.mongodb.com/docs/v5.0/reference/operator/update/pull/
	https://pkg.go.dev/go.mongodb.org/mongo-driver/mongo#example-Collection.UpdateOne
*/
func (m *MongoStore) Pull(ctx context.Context, condition, removeCondition docstore.Field) error {
	/*
		Example
			filter := bson.D{{"menus.id", "PlatformMenuID"}}
			update := bson.D{{"$pull", bson.D{{"menus", bson.D{{"id", "PlatformMenuID"}}}}}}
	*/
	upsert := false
	filter := toMongoE(condition)
	update := bson.D{}
	update = append(update, bson.E{Key: "$pull", Value: toMongoE(removeCondition)})
	res, err := m.store.UpdateOne(ctx, filter, update, &options.UpdateOptions{Upsert: &upsert})
	if err != nil {
		return err
	}

	if res.MatchedCount == 0 {
		return docstore.NotFound
	}

	if res.ModifiedCount == 0 {
		return docstore.NothingUpdated
	}

	return err
}

func convertTime(obj map[string]interface{}) {
	for k, v := range obj {
		if reflect.TypeOf(v) == reflect.TypeOf(time.Time{}) {
			obj[k] = primitive.NewDateTimeFromTime(v.(time.Time))
		}
	}
}

type MongoIterator struct {
	cursor *mongo.Cursor
}

func NewMongoIterator(cursor *mongo.Cursor) *MongoIterator {
	return &MongoIterator{cursor: cursor}
}

func (i *MongoIterator) Next(ctx context.Context, doc interface{}) error {
	if i.cursor.TryNext(ctx) {
		return i.cursor.Decode(doc)
	}

	return docstore.EndOfDoc
}

func (i *MongoIterator) Close(ctx context.Context) error {
	return i.cursor.Close(ctx)
}

func (m *MongoStore) toIndexSpecification(idx map[string]map[string]interface{}) *mongo.IndexSpecification {
	spec := &mongo.IndexSpecification{}
	for k, v := range idx {
		switch k {
		case "keys":
			for f, fv := range v {
				switch value := fv.(type) {
				case int:
					spec.KeysDocument = bson.Raw(bsoncore.NewDocumentBuilder().AppendInt32(f, int32(value)).Build())
					spec.Name = fmt.Sprintf("%s_%s", f, strconv.Itoa(value))
				case []interface{}:
					docBuilder := bsoncore.NewDocumentBuilder()
					name := make([]string, 0, len(value))
					for _, ci := range value {
						mi, ok := ci.(map[string]interface{})
						if !ok {
							continue
						}
						for field, sort := range mi {
							i, ok := sort.(int)
							if ok {
								docBuilder.AppendInt32(field, int32(i))
								name = append(name, fmt.Sprintf("%s_%s", f, strconv.Itoa(i)))
							}
						}
					}
					spec.KeysDocument = bson.Raw(docBuilder.Build())
					spec.Name = strings.Join(name, "_")
				}
			}
		case "options":
			for o, ov := range v {
				switch o {
				case IndexOptionName:
					name, ok := ov.(string)
					if ok {
						spec.Name = name
					}
				case IndexOptionUnique:
					unique, ok := ov.(bool)
					if ok {
						spec.Unique = &unique
					}
				case IndexOptionExpireAfterSeconds:
					seconds, ok := ov.(int32)
					if ok {
						spec.ExpireAfterSeconds = &seconds
					}
				case IndexOptionSparse:
					sparse, ok := ov.(bool)
					if ok {
						spec.Sparse = &sparse
					}
				}
			}
		}

	}

	spec.Namespace = fmt.Sprintf("%s.%s", m.store.Database().Name(), m.collection)
	return spec
}

func isIndexSpecExist(newSpec *mongo.IndexSpecification, specs []*mongo.IndexSpecification) bool {
	for _, currentSpec := range specs {
		if newSpec.Namespace == currentSpec.Namespace &&
			newSpec.KeysDocument.String() == currentSpec.KeysDocument.String() &&
			(newSpec.ExpireAfterSeconds == currentSpec.ExpireAfterSeconds || (newSpec.ExpireAfterSeconds != nil && currentSpec.ExpireAfterSeconds != nil && *newSpec.ExpireAfterSeconds == *currentSpec.ExpireAfterSeconds)) &&
			(newSpec.Sparse == currentSpec.Sparse || (newSpec.Sparse != nil && currentSpec.Sparse != nil && *newSpec.Sparse == *currentSpec.Sparse)) &&
			(newSpec.Unique == currentSpec.Unique || (newSpec.Unique != nil && currentSpec.Unique != nil && *newSpec.Unique == *currentSpec.Unique)) &&
			(newSpec.Clustered == currentSpec.Clustered || (newSpec.Clustered != nil && currentSpec.Clustered != nil && *newSpec.Clustered == *currentSpec.Clustered)) {
			return true
		}
	}
	return false
}

func isNew(newSpec *mongo.IndexSpecification, specs []*mongo.IndexSpecification) bool {
	for _, currentSpec := range specs {
		if newSpec.Namespace == currentSpec.Namespace && newSpec.Name == currentSpec.Name && newSpec.KeysDocument.String() == currentSpec.KeysDocument.String() {
			if notEqualInt32(newSpec.ExpireAfterSeconds, currentSpec.ExpireAfterSeconds) ||
				notEqualBool(newSpec.Sparse, currentSpec.Sparse) ||
				notEqualBool(newSpec.Unique, currentSpec.Unique) ||
				notEqualBool(newSpec.Clustered, currentSpec.Clustered) {
				return true
			}
		}
	}
	return false
}

func notEqualInt32(newExpireAfterSeconds *int32, currentExpireAfterSeconds *int32) bool {
	if newExpireAfterSeconds == currentExpireAfterSeconds {
		return false
	}

	if (newExpireAfterSeconds == nil && currentExpireAfterSeconds != nil) || (newExpireAfterSeconds != nil && currentExpireAfterSeconds == nil) {
		return true
	}

	if newExpireAfterSeconds != nil && currentExpireAfterSeconds != nil && *newExpireAfterSeconds != *currentExpireAfterSeconds {
		return true
	}

	return false
}

func notEqualBool(newProperty *bool, currentProperty *bool) bool {
	if newProperty == currentProperty {
		return false
	}

	if (newProperty == nil && currentProperty != nil) || (newProperty != nil && currentProperty == nil) {
		return true
	}

	if newProperty != nil && currentProperty != nil && *newProperty != *currentProperty {
		return true
	}

	return false
}

func toIndexModel(i map[string]map[string]interface{}) mongo.IndexModel {
	var (
		d       bson.D
		options = &options.IndexOptions{}
	)
	for k, v := range i {
		switch k {
		case "keys":
			for f, fv := range v {
				switch value := fv.(type) {
				case int:
					d = bson.D{
						bson.E{
							Key:   f,
							Value: value,
						},
					}
				case []interface{}:
					for _, ci := range value {
						mi, ok := ci.(map[string]interface{})
						if !ok {
							continue
						}
						for field, sort := range mi {
							d = append(d, bson.E{
								Key:   field,
								Value: sort,
							})
						}
					}
				}
			}
		case "options":
			for o, ov := range v {
				switch o {
				case IndexOptionName:
					name, ok := ov.(string)
					if ok {
						options.SetName(name)
					}
				case IndexOptionUnique:
					unique, ok := ov.(bool)
					if ok {
						options.SetUnique(unique)
					}
				case IndexOptionExpireAfterSeconds:
					seconds, ok := ov.(int32)
					if ok {
						options.SetExpireAfterSeconds(seconds)
					}
				case IndexOptionSparse:
					sparse, ok := ov.(bool)
					if ok {
						options.SetSparse(sparse)
					}
				}
			}
		}

	}

	return mongo.IndexModel{
		Keys:    d,
		Options: options,
	}
}
