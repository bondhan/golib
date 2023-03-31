package firestore

import (
	"context"
	"errors"
	"fmt"
	"os"

	"cloud.google.com/go/firestore"
	"github.com/bondhan/golib/log"
	"google.golang.org/api/iterator"

	"github.com/bondhan/golib/client"
	"github.com/bondhan/golib/constant"
	"github.com/bondhan/golib/util"

	"github.com/bondhan/golib/docstore"
)

type FireStore struct {
	store      *firestore.CollectionRef
	client     *firestore.Client
	idField    string
	collection string
}

func init() {
	docstore.RegisterDriver("firestore", FireStoreFactory)
}

func FireStoreFactory(config *docstore.Config) (docstore.Driver, error) {
	return NewFireStore(config)
}

func NewFireStore(config *docstore.Config) (*FireStore, error) {
	var cl *firestore.Client
	switch con := config.Connection.(type) {
	case *firestore.Client:
		cl = con
	default:
		os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", config.Credential)
		cl = client.FirestoreClient(context.Background(), config.Database)
	}

	return &FireStore{
		store:      cl.Collection(config.Collection),
		idField:    config.IDField,
		collection: config.Collection,
		client:     cl,
	}, nil
}

func (f *FireStore) Create(ctx context.Context, doc interface{}) error {
	id, idErr := f.getID(doc)
	d := make(map[string]interface{})
	if err := util.DecodeJSON(doc, d); err != nil {
		return err
	}
	if idErr != nil {
		_, _, err := f.store.Add(ctx, d)
		return err
	}
	ref := f.store.Doc(fmt.Sprintf("%v", id))
	_, err := ref.Create(ctx, d)
	return err
}

func (f *FireStore) Update(ctx context.Context, id, doc interface{}, replace bool) error {
	ref := f.store.Doc(fmt.Sprintf("%v", id))
	d := make(map[string]interface{})
	if err := util.DecodeJSON(doc, d); err != nil {
		return err
	}
	opt := make([]firestore.SetOption, 0)
	if !replace {
		opt = append(opt, firestore.MergeAll)
	}
	_, err := ref.Set(ctx, d, opt...)
	return err
}

func (f *FireStore) UpdateMany(ctx context.Context, filters []docstore.FilterOpt, fields map[string]interface{}) error {
	return docstore.OperationNotSupported
}

func (f *FireStore) Upsert(ctx context.Context, id, doc interface{}) error {
	ref := f.store.Doc(fmt.Sprintf("%v", id))
	d := make(map[string]interface{})
	if err := util.DecodeJSON(doc, &d); err != nil {
		return err
	}
	opt := []firestore.SetOption{firestore.MergeAll}
	_, err := ref.Set(ctx, d, opt...)
	return err
}

func (f *FireStore) UpdateField(ctx context.Context, id interface{}, fields []docstore.Field) error {
	ref := f.store.Doc(fmt.Sprintf("%v", id))
	ups := make([]firestore.Update, 0)
	for _, f := range fields {
		ups = append(ups, firestore.Update{Path: f.Name, Value: f.Value})
	}
	_, err := ref.Update(ctx, ups)
	return err
}

func (f *FireStore) Increment(ctx context.Context, id interface{}, key string, value int) error {
	ref := f.store.Doc(fmt.Sprintf("%v", id))
	_, err := ref.Update(ctx, []firestore.Update{{Path: key, Value: firestore.Increment(value)}})
	return err
}

func (f *FireStore) GetIncrement(ctx context.Context, id interface{}, key string, value int, doc interface{}) error {
	ref := f.store.Doc(fmt.Sprintf("%v", id))
	_, err := ref.Update(ctx, []firestore.Update{{Path: key, Value: firestore.Increment(value)}})
	if err != nil {
		return err
	}

	return f.Get(ctx, id, doc)
}

func (f *FireStore) Delete(ctx context.Context, id interface{}) error {
	_, err := f.store.Doc(fmt.Sprintf("%v", id)).Delete(ctx)
	return err
}

func (f *FireStore) DeleteMany(ctx context.Context, query *docstore.QueryOpt) error {
	return docstore.OperationNotSupported
}

func (f *FireStore) Get(ctx context.Context, id interface{}, doc interface{}) error {
	ds, err := f.store.Doc(fmt.Sprintf("%v", id)).Get(ctx)
	if err != nil {
		return err
	}
	d := ds.Data()
	d[f.idField] = ds.Ref.ID
	return util.DecodeJSON(d, doc)
}

func (f *FireStore) Count(ctx context.Context, query *docstore.QueryOpt) (int64, error) {
	_, fquery, err := getFireQuery(query, f.store.Query)
	if err != nil {
		return -1, err
	}

	iter := fquery.Snapshots(ctx)
	defer iter.Stop()

	qsnap, err := iter.Next()

	if err != nil {
		return -1, err
	}

	return int64(qsnap.Size), nil
}

func (f *FireStore) Find(ctx context.Context, query *docstore.QueryOpt, docs interface{}) error {
	qs, fquery, err := getFireQuery(query, f.store.Query)
	if err != nil {
		return err
	}

	iter := fquery.Documents(ctx)
	var out []map[string]interface{}
	count := 0
	limit := 0
	skip := 0
	if qs != nil {
		skip = qs.Skip
		limit = skip + qs.Limit
	}

	defer iter.Stop()
	for {
		doc, err := iter.Next()
		if err != nil {
			if err == iterator.Done {
				break
			}
			return err
		}

		tmp := doc.Data()
		tmp[f.idField] = doc.Ref.ID

		if qs != nil {
			if isMatch(tmp, qs) {
				count++
				if skip > 0 && count <= skip {
					continue
				}
				if limit > 0 && count > limit {
					break
				}
				out = append(out, tmp)
			}
			continue
		}

		out = append(out, tmp)
	}

	return util.DecodeJSON(out, docs)
}

func (f *FireStore) FindOne(ctx context.Context, query *docstore.QueryOpt, doc interface{}) error {

	var id interface{}

	for _, p := range query.Filter {
		if p.Field == f.idField && (p.Ops == constant.EQ || p.Ops == constant.SE) {
			id = p.Value
			break
		}
	}

	if id != nil {
		return f.Get(ctx, id, doc)
	}

	query.Limit = 1
	_, fquery, err := getFireQuery(query, f.store.Query)
	if err != nil {
		return err
	}

	iter := fquery.Documents(ctx)

	defer iter.Stop()
	d, err := iter.Next()
	if err != nil {
		if err == iterator.Done {
			return docstore.NotFound
		}
		return err
	}

	tmp := d.Data()
	tmp[f.idField] = d.Ref.ID
	return util.DecodeJSON(tmp, doc)
}

func (f *FireStore) Query(ctx context.Context, query *docstore.QueryOpt) (docstore.Iterator, error) {
	qs, fquery, err := getFireQuery(query, f.store.Query)
	if err != nil {
		return nil, err
	}

	iter := fquery.Documents(ctx)
	return NewFireIterator(iter, qs, f.idField), nil
}

func (f *FireStore) BulkCreate(ctx context.Context, docs []interface{}, opts ...interface{}) error {
	batch := f.client.Batch()

	for _, doc := range docs {
		id, err := f.getID(doc)
		if err != nil {
			continue
		}
		d := make(map[string]interface{})
		if err := util.DecodeJSON(doc, d); err != nil {
			return err
		}
		batch = batch.Create(f.store.Doc(fmt.Sprintf("%v", id)), d)
	}
	_, err := batch.Commit(ctx)
	return err
}

func (f *FireStore) BulkGet(ctx context.Context, ids []interface{}, docs interface{}) error {
	ds := make([]*firestore.DocumentRef, 0)
	for _, i := range ids {
		ds = append(ds, f.store.Doc(fmt.Sprintf("%v", i)))
	}
	snaps, err := f.client.GetAll(ctx, ds)
	if err != nil {
		return err
	}
	var out []map[string]interface{}

	for _, s := range snaps {
		out = append(out, s.Data())
	}

	return util.DecodeJSON(out, docs)
}
func (f *FireStore) Migrate(ctx context.Context, config interface{}) error {
	return nil
}

func (f *FireStore) As(i interface{}) bool {
	p, ok := i.(**firestore.CollectionRef)
	if !ok {
		return false
	}
	*p = f.store
	return true
}

func (f *FireStore) getID(doc interface{}) (interface{}, error) {
	idf := f.idField
	if util.IsStructOrPointerOf(doc) {
		var err error
		idf, err = util.FindFieldByTag(doc, "json", f.idField)
		if err != nil {
			return nil, err
		}
	}

	id, ok := util.Lookup(idf, doc)

	if !ok {
		return nil, errors.New("[docstore/firestore] missing document ID")
	}

	return id, nil
}

func (f *FireStore) Ping(ctx context.Context) error {
	log.GetLogger(ctx, "docstore", "ping").Warn("[docstore/firestore] not implement ping")
	return nil
}

func (f *FireStore) Distinct(ctx context.Context, fieldName string, filter interface{}) ([]interface{}, error) {
	log.GetLogger(ctx, "docstore", "distinct").Warn("[docstore/firestore] not implement distinct")
	return nil, errors.New("[docstore/firestore] not implement distinct")
}

func (f *FireStore) Pull(ctx context.Context, condition, removeCondition docstore.Field) error {
	log.GetLogger(ctx, "docstore", "pull").Warn("[docstore/firestore] not implement pull")
	return nil
}

func (f *FireStore) Disconnect(ctx context.Context) error {
	if f.client != nil {
		return f.client.Close()
	}
	return errors.New("client is nil")
}

type FireIterator struct {
	iter    *firestore.DocumentIterator
	query   *docstore.QueryOpt
	idField string
	limit   int
	skip    int
	count   int
}

func NewFireIterator(iter *firestore.DocumentIterator, query *docstore.QueryOpt, idField string) *FireIterator {
	fi := &FireIterator{
		iter:    iter,
		idField: idField,
	}
	if query != nil {
		fi.query = query
		fi.skip = query.Skip
		fi.limit = query.Limit
		fi.count = 0
	}
	return fi
}

func (i *FireIterator) Next(ctx context.Context, doc interface{}) error {
	if i.query != nil {
		return i.findNext(ctx, doc)
	}

	d, err := i.iter.Next()

	if err != nil {
		if err == iterator.Done {
			return docstore.EndOfDoc
		}
		return err
	}

	tmp := d.Data()
	tmp[i.idField] = d.Ref.ID
	return util.DecodeJSON(tmp, doc)
}

func (i *FireIterator) findNext(ctx context.Context, doc interface{}) error {
	for {
		d, err := i.iter.Next()
		if err != nil {
			if err == iterator.Done {
				return docstore.EndOfDoc
			}
			return err
		}

		tmp := d.Data()
		tmp[i.idField] = d.Ref.ID

		if isMatch(tmp, i.query) {
			i.count++
			if i.skip > 0 && i.count <= i.skip {
				continue
			}
			if i.limit > 0 && i.count > i.limit {
				return docstore.EndOfDoc
			}
			return util.DecodeJSON(tmp, doc)
		}
		continue
	}
}

func (i *FireIterator) Close(ctx context.Context) error {
	i.iter.Stop()
	return nil
}
