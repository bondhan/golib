package docstore

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/bondhan/golib/cache"
	"github.com/bondhan/golib/log"
	"github.com/bondhan/golib/util"
)

const (
	defaultID         = "id"
	defaultTimestamp  = "created_at"
	defaultExpiration = 3600 * 24
)

type Config struct {
	Database          string                              `json:"db,omitempty"`
	Collection        string                              `json:"collection,omitempty"`
	SchemaVersion     string                              `json:"schema_version,omitempty"`
	CacheURL          string                              `json:"cache_url,omitempty"`
	CacheExpiration   int                                 `json:"cache_expiration,omitempty"`
	IDField           string                              `json:"id_field,omitempty"`
	TimestampField    string                              `json:"timestamp_field,omitempty"`
	UpdateTimeField   string                              `json:"updatetime_field,omitempty"`
	Driver            string                              `json:"driver,omitempty"`
	Connection        interface{}                         `json:"connection,omitempty"`
	Credential        string                              `json:"credential,omitempty"`
	Indexes           []map[string]map[string]interface{} `json:"indexes,omitempty"`
	DropExistingIndex bool                                `json:"drop_existing_index"`
	CacheCount        bool                                `json:"cache_count,omitempty"`
	IDGenerator       IDGenerator
	TimeGenerator     TimeGenerator
}

func (c *Config) validate() error {

	if c.Database == "" {
		return errors.New("[docstore] missing database param")
	}

	if c.Collection == "" {
		return errors.New("[docstore] missing collection param")
	}

	if c.CacheURL == "" {
		return errors.New("[docstore] missing cache_url param")
	}

	if c.Driver == "" {
		return errors.New("[docstore] missing driver param")
	}

	if c.Connection == nil {
		return errors.New("[docstore] missing connection param")
	}

	if c.IDField == "" {
		c.IDField = defaultID
	}

	if c.CacheExpiration == 0 {
		c.CacheExpiration = defaultExpiration
	}

	return nil
}

type CachedStore struct {
	*Config
	cache   *cache.Cache
	storage Driver
}

func New(config *Config) (*CachedStore, error) {
	if err := config.validate(); err != nil {
		return nil, err
	}
	dv, err := GetDriver(config)
	if err != nil {
		return nil, err
	}

	cache, err := cache.New(config.CacheURL)
	if err != nil {
		return nil, err
	}

	return NewDocstore(dv, cache, config), nil
}

func NewDocstore(storage Driver, cache *cache.Cache, config *Config) *CachedStore {
	config.IDGenerator = DefaultIDGenerator
	config.TimeGenerator = DefaultTimeGenerator
	return &CachedStore{
		Config:  config,
		cache:   cache,
		storage: storage,
	}
}

func (s *CachedStore) getID(doc interface{}) (interface{}, error) {
	if util.IsMap(doc) {
		id, ok := util.Lookup(s.IDField, doc)

		if !ok {
			return nil, errors.New("[docstore] missing document ID")
		}

		return id, nil
	}

	idf, err := util.FindFieldByTag(doc, "json", s.IDField)
	if err != nil {
		return nil, err
	}

	id, ok := util.Lookup(idf, doc)

	if !ok {
		return nil, errors.New("[docstore] missing document ID")
	}

	return id, nil
}

func (s *CachedStore) setID(doc interface{}, idField string) error {
	if idField == "" {
		return errors.New("[docstore/cached] missing id field param")
	}

	idf := idField

	if util.IsStructOrPointerOf(doc) {
		idfl, err := util.FindFieldByTag(doc, "json", idField)
		if err != nil {
			return err
		}
		idf = idfl
	}

	_, ok := util.Lookup(idf, doc)

	if !ok {
		var idk reflect.Type
		if util.IsMap(doc) {
			idk = reflect.TypeOf("")
		}
		if idk == nil {
			ft, err := util.FindFieldTypeByTag(doc, "json", idField)
			if err != nil {
				return err
			}
			idk = ft
		}

		uid := s.IDGenerator(idk, doc)
		return util.SetValue(doc, idf, uid)

	}

	return nil
}

func (s *CachedStore) setTime(ctx context.Context, doc interface{}, tField string, overwrite bool) error {
	if tField == "" {
		return nil
	}

	tf := tField

	if util.IsStructOrPointerOf(doc) {
		tsf, err := util.FindFieldByTag(doc, "json", tf)
		if err != nil {
			return err
		}
		tf = tsf
	}

	_, ok := util.Lookup(tf, doc)
	if !ok || overwrite {
		var tsk reflect.Type
		if util.IsMap(doc) {
			tsk = reflect.TypeOf(time.Time{})
		}
		if tsk == nil {
			ft, err := util.FindFieldTypeByTag(doc, "json", tField)
			if err != nil {
				return err
			}
			tsk = ft
		}
		util.SetValue(doc, tf, s.TimeGenerator(tsk, doc))
	}
	return nil
}

func (s *CachedStore) Create(ctx context.Context, doc interface{}) error {
	if err := s.setID(doc, s.IDField); err != nil {
		return err
	}

	if err := s.setTime(ctx, doc, s.TimestampField, false); err != nil {
		return err
	}

	if err := s.setTime(ctx, doc, s.UpdateTimeField, false); err != nil {
		return err
	}

	return s.storage.Create(ctx, doc)
}

func (s *CachedStore) update(ctx context.Context, doc interface{}, replace, upsert bool) error {
	id, err := s.getID(doc)
	if err != nil {
		return err
	}

	if s.CacheExpiration != 1 {
		if err := s.cache.Delete(ctx, fmt.Sprintf("%v", id)); err != nil {
			log.GetLogger(ctx, "docstore", "update").WithError(err).Error("error deleting cache ")
		}
	}

	if err := s.setTime(ctx, doc, s.UpdateTimeField, true); err != nil {
		return err
	}

	if upsert {
		return s.storage.Upsert(ctx, id, doc)
	}

	return s.storage.Update(ctx, id, doc, replace)
}

func (s *CachedStore) Update(ctx context.Context, doc interface{}) error {
	return s.update(ctx, doc, false, false)
}

// UpdateMany updates multiple document fields matching the filters
//
// Fields are required to be passed as map[string]interface.
// To update a nested field, pass the key as:
//
//	map[string]interface{}{...
//	"foo.nestedBar" = "baz"
//
// ...}
func (s *CachedStore) UpdateMany(ctx context.Context, filters []FilterOpt, doc map[string]interface{}) error {
	return s.storage.UpdateMany(ctx, filters, doc)
}

func (s *CachedStore) Upsert(ctx context.Context, doc interface{}) error {
	return s.update(ctx, doc, false, true)
}

func (s *CachedStore) UpdateField(ctx context.Context, id interface{}, key string, value interface{}) error {

	if s.CacheExpiration != 1 {
		if err := s.cache.Delete(ctx, fmt.Sprintf("%v", id)); err != nil {
			log.GetLogger(ctx, "docstore", "UpdateField").WithError(err).Error("error deleting cache ")
		}
	}

	return s.storage.UpdateField(ctx, id, []Field{{Name: key, Value: value}})
}

func (s *CachedStore) Pull(ctx context.Context, condition, removeCondition Field) error {
	return s.storage.Pull(ctx, condition, removeCondition)
}

func (s *CachedStore) UpdateFields(ctx context.Context, id interface{}, value []Field) error {

	if s.CacheExpiration != 1 {
		if err := s.cache.Delete(ctx, fmt.Sprintf("%v", id)); err != nil {
			log.GetLogger(ctx, "docstore", "UpdateFields").WithError(err).Error("error deleting cache ")
		}
	}

	return s.storage.UpdateField(ctx, id, value)
}

func (s *CachedStore) Increment(ctx context.Context, id interface{}, fieldName string, value int) error {
	if s.CacheExpiration != 1 {
		if err := s.cache.Delete(ctx, fmt.Sprintf("%v", id)); err != nil {
			log.GetLogger(ctx, "docstore", "Increment").WithError(err).Error("error deleting cache ")
		}
	}

	return s.storage.Increment(ctx, id, fieldName, value)
}

func (s *CachedStore) Replace(ctx context.Context, doc interface{}) error {
	return s.update(ctx, doc, true, false)
}

func (s *CachedStore) Get(ctx context.Context, id, doc interface{}) error {

	if !util.IsPointerOfStruct(doc) && !util.IsMap(doc) {
		return errors.New("[docstore] docs should be a pointer of struct or map")
	}

	if s.CacheExpiration != 1 {
		if s.cache.Exist(ctx, fmt.Sprintf("%v", id)) {
			if err := s.cache.Get(ctx, fmt.Sprintf("%v", id), doc); err == nil {
				return nil
			}
		}
	}

	if err := s.storage.Get(ctx, id, doc); err != nil {
		return err
	}

	if s.CacheExpiration != 1 {
		return s.cache.Set(ctx, fmt.Sprintf("%v", id), doc, s.CacheExpiration)
	}

	return nil
}

func (s *CachedStore) Delete(ctx context.Context, id interface{}) error {
	if s.CacheExpiration != 1 {
		if err := s.cache.Delete(ctx, fmt.Sprintf("%v", id)); err != nil {
			log.GetLogger(ctx, "docstore", "Delete").WithError(err).Error("error deleting cache ")
		}
	}

	return s.storage.Delete(ctx, id)
}

// Delete Many delete documents matching the filters
func (s *CachedStore) DeleteMany(ctx context.Context, query *QueryOpt) error {
	return s.storage.DeleteMany(ctx, query)
}

func (s *CachedStore) Find(ctx context.Context, query *QueryOpt, docs interface{}) error {

	if !util.IsPointerOfSlice(docs) {
		return errors.New("[docstore] docs should be a pointer of slice")
	}

	return s.storage.Find(ctx, query, docs)
}

func (s *CachedStore) Count(ctx context.Context, query *QueryOpt) (int64, error) {
	key := "count:ALL"

	if s.CacheCount {
		if query != nil {
			key = "count:" + query.Hash()
		}
		if c, err := s.cache.GetInt(ctx, key); err == nil {
			return c, nil
		}
	}

	i, err := s.storage.Count(ctx, query)
	if err != nil {
		return -1, err
	}

	if s.CacheCount && i > 0 {
		s.cache.Set(ctx, key, i, s.CacheExpiration)
	}

	return i, nil
}

func (s *CachedStore) FindOne(ctx context.Context, query *QueryOpt, doc interface{}) error {

	if !util.IsPointerOfStruct(doc) && !util.IsMap(doc) {
		return errors.New("[docstore] docs should be a pointer of struct or map")
	}

	return s.storage.FindOne(ctx, query, doc)
}

func (s *CachedStore) IsExists(ctx context.Context, query *QueryOpt) (bool, error) {
	var doc interface{}

	if err := s.storage.FindOne(ctx, query, &doc); err != nil {
		if err == NotFound {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (s *CachedStore) BulkCreate(ctx context.Context, docs interface{}, opts ...interface{}) error {
	if !util.IsSlice(docs) {
		return errors.New("[docstore] documents should be a slice")
	}

	rdocs := reflect.ValueOf(docs)
	ins := make([]interface{}, rdocs.Len())
	for i := 0; i < rdocs.Len(); i++ {
		ins[i] = rdocs.Index(i).Interface()
	}

	for i, d := range ins {

		if err := s.setID(d, s.IDField); err != nil {
			return err
		}

		if err := s.setTime(ctx, d, s.TimestampField, false); err != nil {
			return err
		}

		if err := s.setTime(ctx, d, s.UpdateTimeField, false); err != nil {
			return err
		}

		ins[i] = d
	}

	return s.storage.BulkCreate(ctx, ins, opts...)
}

func (s *CachedStore) BulkGet(ctx context.Context, ids, docs interface{}) error {
	if !util.IsSlice(ids) {
		return errors.New("[docstore] IDs should be a slice")
	}

	if !util.IsPointerOfSlice(docs) {
		return errors.New("[docstore] docs should be a pointer of slice")
	}

	rids := reflect.ValueOf(ids)
	ins := make([]interface{}, rids.Len())
	for i := 0; i < rids.Len(); i++ {
		ins[i] = rids.Index(i).Interface()
	}

	return s.storage.BulkGet(ctx, ins, docs)
}

func (s *CachedStore) Migrate(ctx context.Context, config interface{}) error {
	return s.storage.Migrate(ctx, config)
}

func (s *CachedStore) GetCache() *cache.Cache {
	return s.cache
}

func (s *CachedStore) GetDriver() Driver {
	return s.storage
}

func (s *CachedStore) Ping(ctx context.Context) error {
	return s.storage.Ping(ctx)
}

func (s *CachedStore) Distinct(ctx context.Context, fieldName string, filter interface{}) ([]interface{}, error) {
	return s.storage.Distinct(ctx, fieldName, filter)
}

type IDGenerator func(reflect.Type, interface{}) interface{}
type TimeGenerator func(reflect.Type, interface{}) interface{}

func DefaultIDGenerator(kind reflect.Type, doc interface{}) interface{} {
	switch kind {
	case reflect.TypeOf(""):
		return util.Hash58(doc)
	case reflect.TypeOf(int64(0)):
		return util.GenerateRandUID()
	case reflect.TypeOf(int(0)):
		return int(util.GenerateRandUID())
	case reflect.TypeOf(uint64(0)):
		return uint64(util.GenerateRandUID())
	case reflect.TypeOf(uint(0)):
		return uint(util.GenerateRandUID())
	default:
		return errors.New("[docstore] unsupported ID type")
	}
}

func DefaultTimeGenerator(kind reflect.Type, doc interface{}) interface{} {
	switch kind {
	case reflect.TypeOf(""):
		return time.Now().String()
	case reflect.TypeOf(int(0)):
		return int(time.Now().Unix())
	case reflect.TypeOf(int64(0)):
		return time.Now().UnixMilli()
	case reflect.TypeOf(uint(0)):
		return uint(time.Now().UnixMicro())
	case reflect.TypeOf(uint64(0)):
		return uint64(time.Now().UnixNano())
	default:
		return time.Now()
	}
}
