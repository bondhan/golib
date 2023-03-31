package docstore

import (
	"context"
	"errors"
	"sort"
	"sync"

	"github.com/bondhan/golib/constant"
	"github.com/bondhan/golib/log"
	"github.com/bondhan/golib/util"
	"github.com/imdario/mergo"
	"go.opentelemetry.io/otel"
)

// var memstore = make(map[string]*MemoryStore)

type MemoryStore struct {
	storage map[interface{}]map[string]interface{}
	idField string
	mux     *sync.Mutex
}

func MemoryStoreFactory(config *Config) (Driver, error) {
	return NewMemoryStore(config.Collection, config.IDField), nil
}

func NewMemoryStore(name, idField string) *MemoryStore {

	m := &MemoryStore{
		storage: make(map[interface{}]map[string]interface{}),
		idField: idField,
		mux:     &sync.Mutex{},
	}

	return m
}

func (m *MemoryStore) getID(doc interface{}) (interface{}, error) {
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
		return nil, errors.New("[docstore/memory] missing document ID")
	}

	return id, nil
}

func (m *MemoryStore) Create(ctx context.Context, doc interface{}) error {
	tracer := otel.Tracer("docstore/memory")
	_, span := tracer.Start(ctx, "Create")
	defer span.End()
	m.mux.Lock()
	defer m.mux.Unlock()

	id, err := m.getID(doc)
	if err != nil {
		return err
	}

	if _, ok := m.storage[id]; ok {
		return errors.New("[docstore/memory] document ID is already exist")
	}

	d := make(map[string]interface{})
	if err := util.DecodeJSON(doc, &d); err != nil {
		return err
	}

	m.storage[id] = d
	return nil
}

func (m *MemoryStore) Update(ctx context.Context, id, doc interface{}, replace bool) error {
	tracer := otel.Tracer("docstore/memory")
	_, span := tracer.Start(ctx, "Update")
	defer span.End()
	m.mux.Lock()
	defer m.mux.Unlock()

	if _, ok := m.storage[id]; !ok {
		return NotFound
	}
	d := make(map[string]interface{})
	if err := util.DecodeJSON(doc, &d); err != nil {
		return err
	}

	if replace {
		m.storage[id] = d
		return nil
	}

	cd := m.storage[id]

	if err := mergo.MergeWithOverwrite(&cd, d); err != nil {
		return err
	}

	m.storage[id] = cd

	return nil
}

func (m *MemoryStore) UpdateMany(ctx context.Context, filters []FilterOpt, fields map[string]interface{}) error {
	return OperationNotSupported
}

func (m *MemoryStore) Upsert(ctx context.Context, id, doc interface{}) error {
	m.mux.Lock()
	_, ok := m.storage[id]
	m.mux.Unlock()
	if !ok {
		return m.Create(ctx, doc)
	}
	return m.Update(ctx, id, doc, false)
}

func (m *MemoryStore) UpdateField(ctx context.Context, id interface{}, fields []Field) error {
	tracer := otel.Tracer("docstore/memory")
	_, span := tracer.Start(ctx, "UpdateField")
	defer span.End()
	m.mux.Lock()
	defer m.mux.Unlock()
	d, ok := m.storage[id]
	if !ok {
		return NotFound
	}

	for _, f := range fields {
		// fn, _ := util.FindFieldByTag(d, "json", f.Name)
		if err := util.SetValue(d, f.Name, f.Value); err != nil {
			return err
		}
	}

	m.storage[id] = d
	return nil
}

func (m *MemoryStore) Pull(ctx context.Context, condition, removeCondition Field) error {
	log.GetLogger(ctx, "docstore", "pull").Warn("[docstore/memory] not implement pull")
	return nil
}

func (m *MemoryStore) Increment(ctx context.Context, id interface{}, key string, value int) error {
	tracer := otel.Tracer("docstore/memory")
	_, span := tracer.Start(ctx, "Increment")
	defer span.End()
	m.mux.Lock()
	defer m.mux.Unlock()
	d, ok := m.storage[id]
	if !ok {
		m.storage[id] = map[string]interface{}{key: value}
		return nil
	}

	field, ok := d[key]
	if !ok {
		return errors.New("[docstore/memory] field not found")
	}

	switch v := field.(type) {
	case int:
		d[key] = v + value
	case int32:
		d[key] = v + int32(value)
	case int64:
		d[key] = v + int64(value)
	case float64:
		d[key] = v + float64(value)
	case float32:
		d[key] = v + float32(value)
	case uint:
		d[key] = v + uint(value)
	case uint32:
		d[key] = v + uint32(value)
	case uint64:
		d[key] = v + uint64(value)
	default:
		return errors.New("[docstore/memory] destination type is not a number")
	}

	m.storage[id] = d
	return nil
}

func (m *MemoryStore) GetIncrement(ctx context.Context, id interface{}, key string, value int, doc interface{}) error {
	tracer := otel.Tracer("docstore/memory")
	_, span := tracer.Start(ctx, "GetIncrement")
	defer span.End()
	if err := m.Increment(ctx, id, key, value); err != nil {
		return err
	}
	return m.Get(ctx, id, doc)
}

func (m *MemoryStore) Delete(ctx context.Context, id interface{}) error {
	tracer := otel.Tracer("docstore/memory")
	_, span := tracer.Start(ctx, "Delete")
	defer span.End()
	m.mux.Lock()
	defer m.mux.Unlock()

	delete(m.storage, id)
	return nil
}

func (m *MemoryStore) DeleteMany(ctx context.Context, query *QueryOpt) error {
	return OperationNotSupported
}

func (m *MemoryStore) Get(ctx context.Context, id interface{}, doc interface{}) error {
	tracer := otel.Tracer("docstore/memory")
	_, span := tracer.Start(ctx, "Get")
	defer span.End()
	m.mux.Lock()
	defer m.mux.Unlock()
	d, ok := m.storage[id]
	if !ok {
		return NotFound
	}

	if err := util.DecodeJSON(d, doc); err != nil {
		return err
	}

	return nil
}

func (m *MemoryStore) Count(ctx context.Context, query *QueryOpt) (int64, error) {
	out := make([]interface{}, 0)
	return m.find(ctx, query, &out)
}

func (m *MemoryStore) Find(ctx context.Context, query *QueryOpt, docs interface{}) error {
	_, err := m.find(ctx, query, docs)
	return err
}

func (m *MemoryStore) find(ctx context.Context, query *QueryOpt, docs interface{}) (int64, error) {
	tracer := otel.Tracer("docstore/memory")
	_, span := tracer.Start(ctx, "Find")
	defer span.End()
	m.mux.Lock()
	defer m.mux.Unlock()
	out := make([]interface{}, 0)
	if query == nil {
		for _, v := range m.storage {
			out = append(out, v)
		}
		if err := util.DecodeJSON(out, docs); err != nil {
			return -1, err
		}
		return int64(len(out)), nil
	}

	for _, d := range m.storage {
		match := false
		if len(query.Filter) == 0 {
			match = true
		}
		for _, f := range query.Filter {
			fn, err := util.FindFieldByTag(d, "json", f.Field)
			if err == nil {
				f.Field = fn
			}
			if !assertVal(f, d) {
				match = false
				break
			}
			match = true
		}
		if match {
			out = append(out, d)
		}
	}

	if query.OrderBy != "" {
		sort.Slice(out, func(i, j int) bool {
			of := query.OrderBy
			if util.IsStructOrPointerOf(out[j]) {
				if f, err := util.FindFieldByTag(out[j], "json", query.OrderBy); err == nil {
					of = f
				}
			}
			val, ok := util.Lookup(of, out[j])
			if ok {
				return util.Assert(of, out[i], val, constant.GT)
			}
			return false
		})
		if query.IsAscend {
			util.Reverse(out)
		}
	}

	if query.Page > 0 && query.Limit > 0 {
		query.Skip = query.Page * query.Limit
	}

	if query.Skip > 0 && len(out) > query.Skip {
		out = out[query.Skip:]
	}

	if query.Limit > 0 && len(out) > query.Limit {
		out = out[:query.Limit]
	}

	if err := util.DecodeJSON(out, docs); err != nil {
		return -1, err
	}
	return int64(len(out)), nil
}

func (m *MemoryStore) FindOne(ctx context.Context, query *QueryOpt, doc interface{}) error {
	tracer := otel.Tracer("docstore/memory")
	_, span := tracer.Start(ctx, "FindOne")
	defer span.End()

	if query != nil && query.OrderBy != "" {
		tmp := make([]interface{}, 0)
		if _, err := m.find(ctx, query, &tmp); err != nil {
			return err
		}
		if len(tmp) > 0 {
			return util.DecodeJSON(tmp[0], doc)
		}
	}

	m.mux.Lock()
	defer m.mux.Unlock()
	for _, d := range m.storage {
		if query == nil || len(query.Filter) == 0 {
			return util.DecodeJSON(d, doc)
		}
		match := false
		for _, f := range query.Filter {
			fn, err := util.FindFieldByTag(d, "json", f.Field)
			if err == nil {
				f.Field = fn
			}
			if !assertVal(f, d) {
				match = false
				break
			}
			match = true
		}

		if match {
			return util.DecodeJSON(d, doc)
		}
	}

	return NotFound

}

func (m *MemoryStore) Query(ctx context.Context, query *QueryOpt) (Iterator, error) {
	return NewMemIterator(m, query), nil
}

func (m *MemoryStore) BulkCreate(ctx context.Context, docs []interface{}, opts ...interface{}) error {
	tracer := otel.Tracer("docstore/memory")
	_, span := tracer.Start(ctx, "BulkCreate")
	defer span.End()
	m.mux.Lock()
	defer m.mux.Unlock()

	for _, doc := range docs {
		id, err := m.getID(doc)
		if err != nil {
			return err
		}

		if _, ok := m.storage[id]; ok {
			return errors.New("[docstore/memory] document ID is already exist")
		}

		switch d := doc.(type) {
		case map[string]interface{}:
			m.storage[id] = d
			continue
		}

		d := make(map[string]interface{})
		if err := util.DecodeJSON(doc, d); err != nil {
			return err
		}

		m.storage[id] = d
	}

	return nil
}

func (m *MemoryStore) BulkGet(ctx context.Context, ids []interface{}, docs interface{}) error {
	tracer := otel.Tracer("docstore/memory")
	_, span := tracer.Start(ctx, "BulkGet")
	defer span.End()
	m.mux.Lock()
	defer m.mux.Unlock()

	out := make([]map[string]interface{}, 0)
	for _, id := range ids {
		d, ok := m.storage[id]
		if !ok {
			return NotFound
		}
		out = append(out, d)
	}

	if err := util.DecodeJSON(out, docs); err != nil {
		return err
	}

	return nil
}

func (m *MemoryStore) Migrate(ctx context.Context, config interface{}) error {
	return nil
}

func (m *MemoryStore) As(i interface{}) bool { return false }

func (m *MemoryStore) Ping(ctx context.Context) error {
	log.GetLogger(ctx, "docstore", "ping").Warn("[docstore/memory] not implement ping")
	return nil
}

func (m *MemoryStore) Distinct(ctx context.Context, fieldName string, filter interface{}) ([]interface{}, error) {
	log.GetLogger(ctx, "docstore", "distinct").Warn("[docstore/memory] not implement distinct")
	return nil, errors.New("[docstore/memory] not implement distinct")
}

func (m *MemoryStore) Disconnect(ctx context.Context) error {
	return nil
}

type MemIterator struct {
	store  *MemoryStore
	query  *QueryOpt
	index  int
	length int
	keys   []interface{}
}

func NewMemIterator(store *MemoryStore, query *QueryOpt) *MemIterator {
	// l := len(store.storage)
	keys := make([]interface{}, 0)
	if query != nil && query.OrderBy != "" {
		tmp := make([]interface{}, 0)
		_, err := store.find(context.Background(), query, &tmp)
		if err == nil {
			for _, t := range tmp {
				id, err := store.getID(t)
				if err == nil {
					keys = append(keys, id)
				}
			}
		}
	} else {
		for k := range store.storage {
			keys = append(keys, k)
		}
	}

	return &MemIterator{
		store:  store,
		query:  query,
		length: len(keys),
		keys:   keys,
		index:  0,
	}
}

func (i *MemIterator) Next(ctx context.Context, doc interface{}) error {
	if i.index >= i.length {
		return EndOfDoc
	}

	if i.query == nil {
		if err := util.DecodeJSON(i.store.storage[i.keys[i.index]], doc); err != nil {
			return err
		}
		i.index++
		return nil
	}

	for n := i.index; n < i.length; n++ {
		d := i.store.storage[i.keys[i.index]]
		i.index++
		match := false
		for _, f := range i.query.Filter {
			fn, err := util.FindFieldByTag(d, "json", f.Field)
			if err == nil {
				f.Field = fn
			}
			if !assertVal(f, d) {
				match = false
				break
			}
			match = true
		}
		if match || len(i.query.Filter) == 0 {
			return util.DecodeJSON(d, doc)
		}
	}

	i.index = i.length
	return EndOfDoc
}

func (i *MemIterator) Close(ctx context.Context) error {
	i.index = i.length
	return nil
}

func assertOr(filter interface{}, ctx interface{}) bool {
	opts, ok := filter.([]FilterOpt)
	if !ok {
		return false
	}

	for _, f := range opts {
		if assertVal(f, ctx) {
			return true
		}
	}
	return false
}

func assertVal(filter FilterOpt, ctx interface{}) bool {
	switch filter.Ops {
	case constant.OR:
		return assertOr(filter.Value, ctx)
	case constant.AND:
		opts, ok := filter.Value.([]FilterOpt)
		if !ok {
			return false
		}
		match := false
		for _, f := range opts {
			if !assertVal(f, ctx) {
				match = false
				break
			}
			match = true
		}
		return match
	case constant.AIN, constant.AM:
		filter.Ops = constant.IN
		fallthrough
	default:
		return util.Assert(filter.Field, ctx, filter.Value, filter.Ops)
	}
}
