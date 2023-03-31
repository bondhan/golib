package mongo

import (
	"context"
	"fmt"
	"os"
	"testing"

	_ "github.com/bondhan/golib/cache/mem"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/bondhan/golib/client"

	"github.com/bondhan/golib/docstore"
)

const (
	mongoURIEnv = "MONGO_URI"
)

func initDriver(t *testing.T) *mongo.Database {
	mongoURI := os.Getenv(mongoURIEnv)
	if mongoURI == "" {
		t.Skipf("set %s to run this test", mongoURIEnv)
	}

	t.Helper()

	client := &client.MongoClient{
		URI:     mongoURI,
		AppName: "test",
	}

	ctx := context.Background()

	cl, err := client.MongoConnect()
	require.Nil(t, err)
	db := cl.Database("test")
	require.NotNil(t, db)

	col := db.Collection("docstore")
	col.DeleteMany(ctx, bson.D{})
	return db
}

func TestIndexCreation(t *testing.T) {
	ctx := context.Background()
	db := initDriver(t)
	ms, err := NewMongostore(db, "userprovider", "id")
	require.Nil(t, err)
	tests := []struct {
		name    string
		indexes []map[string]map[string]interface{}
		want    int
	}{
		{
			name: "create new",
			indexes: []map[string]map[string]interface{}{
				{
					"keys": {
						"name": IndexTypeSortOrderAscending,
					},
				},
				{
					"keys": {
						"user": IndexTypeSortOrderAscending,
					},
					"options": {
						IndexOptionUnique: true,
					},
				},
			},
			want: 6,
		},
		{
			name: "drop old",
			indexes: []map[string]map[string]interface{}{
				{
					"keys": {
						"user": IndexTypeSortOrderAscending,
					},
					"options": {
						IndexOptionUnique: true,
					},
				},
			},
			want: 5,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbconf := &docstore.Config{
				IDField:           "id",
				TimestampField:    "createdAt",
				UpdateTimeField:   "updatedAt",
				Indexes:           tt.indexes,
				DropExistingIndex: true,
			}
			require.Nil(t, ms.Migrate(ctx, dbconf))
			idxs, err := ms.listIndex(ctx)
			require.NoError(t, err)
			assert.Equal(t, tt.want, len(idxs))
		})
	}
}

func TestMongoStore(t *testing.T) {
	db := initDriver(t)

	ms, err := NewMongostore(db, "docstore", "id")
	ms.Migrate(context.Background(), nil)

	require.Nil(t, err)
	docstore.DriverCRUDTest(ms, t)
	docstore.DriverBulkTest(ms, t)
}

func TestDocstore(t *testing.T) {
	docstore.RegisterDriver("mongo", MongoStoreFactory)

	db := initDriver(t)
	config := &docstore.Config{
		Database:       "test",
		Collection:     "docstore",
		IDField:        "id",
		Driver:         "mongo",
		Connection:     db,
		CacheURL:       "mem://ms",
		TimestampField: "created_at",
	}

	cs, err := docstore.New(config)

	require.Nil(t, err)
	require.NotNil(t, cs)

	cs.Migrate(context.Background(), nil)

	docstore.DocstoreTestCRUD(cs, t)
}

func TestMongoStore_Ping(t *testing.T) {
	db := initDriver(t)
	ctx := context.Background()

	type fields struct {
		store      *mongo.Collection
		idField    string
		collection string
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "ping success test",
			fields: fields{
				store:      db.Collection("test"),
				idField:    "_id",
				collection: "test",
			},
			args: args{
				ctx: ctx,
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MongoStore{
				store:      tt.fields.store,
				idField:    tt.fields.idField,
				collection: tt.fields.collection,
			}

			tt.wantErr(t, m.Ping(tt.args.ctx), fmt.Sprintf("Ping(%v)", tt.args.ctx))
		})
	}
}

func TestMongoStore_Distinct(t *testing.T) {
	db := initDriver(t)
	ctx := context.Background()
	type fields struct {
		store      *mongo.Collection
		idField    string
		collection string
	}
	type args struct {
		ctx       context.Context
		fieldName string
		filter    interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []interface{}
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "distinct success test",
			fields: fields{
				store:      db.Collection("test"),
				idField:    "_id",
				collection: "test",
			},
			args: args{
				ctx:       ctx,
				fieldName: "name",
				filter:    map[string]interface{}{},
			},
			wantErr: assert.NoError,
			want:    []interface{}{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MongoStore{
				store:      tt.fields.store,
				idField:    tt.fields.idField,
				collection: tt.fields.collection,
			}
			got, err := m.Distinct(tt.args.ctx, tt.args.fieldName, tt.args.filter)
			if !tt.wantErr(t, err, fmt.Sprintf("Distinct(%v, %v, %v)", tt.args.ctx, tt.args.fieldName, tt.args.filter)) {
				return
			}
			assert.Equalf(t, tt.want, got, "Distinct(%v, %v, %v)", tt.args.ctx, tt.args.fieldName, tt.args.filter)
		})
	}
}
