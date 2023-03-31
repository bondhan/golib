package retailer

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/bondhan/golib/client"
	mongodoc "github.com/bondhan/golib/docstore/mongo"
	domain "github.com/bondhan/golib/domain/retailer"
	"github.com/bondhan/golib/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
)

const mongoURIEnv = "MONGO_URI"

func TestQuery(t *testing.T) {
	mongoURI := os.Getenv(mongoURIEnv)
	//mongoURI := "mongodb://localhost:27017"
	if mongoURI == "" {
		t.Skipf("set %s to run this test", mongoURIEnv)
	}

	client := &client.MongoClient{
		URI:     mongoURI,
		AppName: "test",
	}

	cl, err := client.MongoConnect()
	require.Nil(t, err)
	db := cl.Database("test")
	require.NotNil(t, db)

	ms, err := mongodoc.NewMongostore(db, "test", "id")
	assert.Nil(t, err)
	assert.NotNil(t, ms)

	require.Nil(t, populateData(ms))

	retailer := &domain.RetailerContext{
		Area: &domain.Area{
			Region:    "Bali",
			RegionSub: "Bali",
			Cluster:   "Denpasar",
		},
		Segments: []string{"segment1", "segment3"},
	}

	builder := NewQueryBuilder(retailer, "*")
	query := builder.WithRegion("area.region").WithSubRegion("area.regionSub").WithCluster("area.cluster").WithActiveUnixTimes("activeTimes", "start", "end").Sort("score", false).Get()
	ctx := context.Background()
	out := make([]map[string]interface{}, 0)
	require.Nil(t, ms.Find(ctx, query, &out))
	assert.Equal(t, 2, len(out))

	assert.Equal(t, int32(30), out[0]["score"])

	//

	retailer.Area = &domain.Area{
		Region:    "Jateng",
		RegionSub: "Jogja",
		Cluster:   "Giwangan",
	}

	builder = NewQueryBuilder(retailer, "*")
	query = builder.WithRegion("area.region").WithSubRegion("area.regionSub").WithCluster("area.cluster").WithActiveUnixTimes("activeTimes", "start", "end").Sort("score", false).Get()
	out = make([]map[string]interface{}, 0)
	require.Nil(t, ms.Find(ctx, query, &out))
	assert.Equal(t, 2, len(out))

	assert.Equal(t, int32(30), out[0]["score"])

	retailer.Area = &domain.Area{
		Region:    "Jateng",
		RegionSub: "Solo",
		Cluster:   "Sukoharjo",
	}

	builder = NewQueryBuilder(retailer, "*")
	query = builder.WithRegion("area.region").WithSubRegion("area.regionSub").WithCluster("area.cluster").WithActiveUnixTimes("activeTimes", "start", "end").Sort("score", false).Get()
	out = make([]map[string]interface{}, 0)
	require.Nil(t, ms.Find(ctx, query, &out))
	assert.Equal(t, 1, len(out))

	assert.Equal(t, int32(0), out[0]["score"])

	db.Collection("test").DeleteMany(ctx, bson.D{})
}

func populateData(ms *mongodoc.MongoStore) error {
	ctx := context.Background()
	docs := []map[string]interface{}{
		{
			"area": map[string]interface{}{
				"region":    []string{"Bali"},
				"regionSub": []string{"*"},
				"cluster":   []string{"Denpasar"},
			},
			"segments": []string{"segment1", "segment2"},
			"activeTimes": []map[string]interface{}{
				{"start": time.Now().Unix(), "end": time.Now().Add(1 * time.Hour).Unix()},
			},
			"score": 30,
		},
		{
			"area": map[string]interface{}{
				"region":    []string{"*"},
				"regionSub": []string{"*"},
				"cluster":   []string{"Giwangan"},
			},
			"segments": []string{"segment1", "segment2"},
			"activeTimes": []map[string]interface{}{
				{"start": time.Now().Unix(), "end": time.Now().Add(1 * time.Hour).Unix()},
			},
			"score": 30,
		},
		{
			"area": map[string]interface{}{
				"region":    []string{"*"},
				"regionSub": []string{"*"},
				"cluster":   []string{"*"},
			},
			"segments": []string{"*"},
			"activeTimes": []map[string]interface{}{
				{"start": time.Now().Unix(), "end": time.Now().Add(1 * time.Hour).Unix()},
			},
			"score": 0,
		},
		{
			"area": map[string]interface{}{
				"region":    []string{"*"},
				"regionSub": []string{"*"},
				"cluster":   []string{"*"},
			},
			"segments": []string{"*"},
			"activeTimes": []map[string]interface{}{
				{"start": time.Now().Add(-1 * time.Hour).Unix(), "end": time.Now().Add(-1 * time.Minute).Unix()},
			},
			"score": 0,
		},
	}

	for _, d := range docs {
		d["id"] = util.Hash58(d)
		if err := ms.Create(ctx, d); err != nil {
			return err
		}
	}
	return nil
}
