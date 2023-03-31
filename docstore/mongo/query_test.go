package mongo

import (
	"context"
	"os"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/bondhan/golib/client"
	"github.com/bondhan/golib/constant"
	"github.com/bondhan/golib/docstore"
)

func Test_toMongoE(t *testing.T) {
	key := "menus.id"
	value := "p1m1s1"
	fields := docstore.Field{
		Name:  key,
		Value: value,
	}
	d := bson.D{}
	d = append(d, bson.E{Key: fields.Name, Value: fields.Value})
	res := toMongoE(fields)
	if !reflect.DeepEqual(res, d) {
		t.Errorf("1 -> toMongoNested() = %v, want %v", res, d)
	}

	// Test 2. Nested
	key2 := "menus"
	value2 := docstore.Field{
		Name:  "id",
		Value: "p1m1s1",
	}

	fields2 := docstore.Field{
		Name:  key2,
		Value: value2,
	}
	d2 := bson.D{}
	d2 = append(d2, bson.E{Key: fields2.Name, Value: bson.D{bson.E{
		Key:   value2.Name,
		Value: value2.Value,
	}},
	})

	res2 := toMongoE(fields2)
	if !reflect.DeepEqual(res2, d2) {
		t.Errorf("2 -> toMongoNested() res2 = %v, want %v", res2, d2)
	}

}

func TestElemMatch(t *testing.T) {
	mongoURI := os.Getenv(mongoURIEnv)
	//mongoURI := "mongodb://localhost:27017"
	if mongoURI == "" {
		t.Skipf("set %s to run this test", mongoURIEnv)
	}

	client := &client.MongoClient{
		URI:     mongoURI,
		AppName: "test",
	}

	ctx := context.Background()

	cl, err := client.MongoConnect()
	require.Nil(t, err)
	db := cl.Database("test")
	require.NotNil(t, db)

	ms, err := NewMongostore(db, "test", "id")
	assert.Nil(t, err)
	assert.NotNil(t, ms)

	doc := map[string]interface{}{
		"id": "1234345434534534",
		"validDates": []map[string]interface{}{
			{"start": 1, "end": 5},
			{"start": 11, "end": 15},
		},
	}

	require.Nil(t, ms.Create(ctx, doc))

	query := docstore.QueryOpt{
		Filter: []docstore.FilterOpt{
			{Field: "validDates", Ops: constant.EM, Value: []docstore.FilterOpt{
				{Field: "start", Ops: constant.LE, Value: 4},
				{Field: "end", Ops: constant.GE, Value: 4},
			}},
		},
	}

	var out map[string]interface{}

	assert.Nil(t, ms.FindOne(ctx, &query, &out))
	assert.Equal(t, "1234345434534534", out["id"])

	db.Collection("test").DeleteMany(ctx, bson.D{})
}
