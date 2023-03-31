//go:build integration

package firestore

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bondhan/golib/docstore"
)

const emulatorHost = "localhost:8083"

// gcloud beta emulators firestore start
func TestFireStore(t *testing.T) {
	os.Setenv("FIRESTORE_EMULATOR_HOST", emulatorHost)
	// t.Skip()
	config := &docstore.Config{
		Database:   "my-project",
		Collection: "docstore",
		IDField:    "id",
		Credential: "credential.json",
	}

	fs, err := NewFireStore(config)
	fs.Migrate(context.Background(), nil)

	iter := fs.store.DocumentRefs(context.Background())
	for {
		doc, err := iter.Next()
		if err != nil {
			break
		}
		doc.Delete(context.Background())
	}

	require.Nil(t, err)
	docstore.DriverCRUDTest(fs, t)
	docstore.DriverBulkTest(fs, t)

}

func TestDocstore(t *testing.T) {
	os.Setenv("FIRESTORE_EMULATOR_HOST", emulatorHost)
	// t.Skip()
	docstore.RegisterDriver("firestore", FireStoreFactory)

	config := &docstore.Config{
		Database:       "my-project",
		Collection:     "docstore",
		IDField:        "id",
		Driver:         "firestore",
		CacheURL:       "mem://ms",
		Credential:     "credential.json",
		Connection:     "",
		TimestampField: "created_at",
	}

	cs, err := docstore.New(config)
	require.Nil(t, err)
	require.NotNil(t, cs)

	cs.Migrate(context.Background(), nil)

	iter := cs.GetDriver().(*FireStore).store.DocumentRefs(context.Background())

	for {
		doc, err := iter.Next()
		if err != nil {
			break
		}
		doc.Delete(context.Background())
	}

	docstore.DocstoreTestCRUD(cs, t)
}

func TestUpsert(t *testing.T) {
	os.Setenv("FIRESTORE_EMULATOR_HOST", emulatorHost)
	// t.Skip()
	docstore.RegisterDriver("firestore", FireStoreFactory)

	config := &docstore.Config{
		Database:       "my-project",
		Collection:     "docstore",
		IDField:        "id",
		Driver:         "firestore",
		CacheURL:       "mem://ms",
		Credential:     "credential.json",
		Connection:     "",
		TimestampField: "created_at",
	}

	cs, err := docstore.New(config)
	require.Nil(t, err)
	require.NotNil(t, cs)

	cs.Migrate(context.Background(), nil)

	iter := cs.GetDriver().(*FireStore).store.DocumentRefs(context.Background())

	for {
		doc, err := iter.Next()
		if err != nil {
			break
		}
		doc.Delete(context.Background())
	}

	ctx := context.Background()
	ts := time.Now()
	usr := map[string]interface{}{
		"id":        "12345",
		"name":      "sahal",
		"username":  "sahalzain",
		"createdAt": ts,
	}

	assert.Nil(t, cs.Upsert(ctx, usr))
	// assert.False(t, true)
}

func TestFireStore_Ping(t *testing.T) {
	type fields struct {
		store      *firestore.CollectionRef
		client     *firestore.Client
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
			name:    "assert error test",
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &FireStore{
				store:      tt.fields.store,
				client:     tt.fields.client,
				idField:    tt.fields.idField,
				collection: tt.fields.collection,
			}
			tt.wantErr(t, f.Ping(tt.args.ctx), fmt.Sprintf("Ping(%v)", tt.args.ctx))
		})
	}
}

func TestFireStore_Distinct(t *testing.T) {
	type fields struct {
		store      *firestore.CollectionRef
		client     *firestore.Client
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
		wantErr error
	}{
		{
			name:    "assert not error test",
			wantErr: errors.New("[docstore/firestore] not implement distinct"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &FireStore{
				store:      tt.fields.store,
				client:     tt.fields.client,
				idField:    tt.fields.idField,
				collection: tt.fields.collection,
			}
			got, err := f.Distinct(tt.args.ctx, tt.args.fieldName, tt.args.filter)
			if err != nil {
				assert.Equal(t, tt.wantErr, err)
				return
			}
			assert.Equalf(t, tt.want, got, "Distinct(%v, %v, %v)", tt.args.ctx, tt.args.fieldName, tt.args.filter)
		})
	}
}
