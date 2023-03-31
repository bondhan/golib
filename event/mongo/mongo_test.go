//go:build integration

package mongo

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/bondhan/golib/client"
	"github.com/bondhan/golib/event"
)

func TestMongoSender(t *testing.T) {
	event.RegisterSender("mongo", NewMongoSender)

	mconf := client.MongoClient{
		URI:     "mongodb://localhost:27017",
		AppName: "test",
	}
	cl, err := mconf.MongoConnect()
	require.Nil(t, err)

	db := cl.Database(mconf.AppName)

	conf := &event.EmitterConfig{
		Sender: &event.DriverConfig{
			Type: "mongo",
			Config: map[string]interface{}{
				"collection": "outbox",
				"connection": db,
			},
		},
	}

	ctx := context.Background()

	em, err := event.New(ctx, conf)
	require.Nil(t, err)
	require.NotNil(t, em)

	key := fmt.Sprintf("%v", time.Now().Unix())

	require.Nil(t, em.Publish(ctx, "test", key, "testdata", nil))

	var out MongoOutbox
	err = db.Collection("outbox").FindOne(ctx, bson.M{"key": key}).Decode(&out)
	require.Nil(t, err)
	assert.Equal(t, key, out.Key)
	db.Collection("outbox").DeleteMany(ctx, bson.D{})
}

func TestMongoWriter(t *testing.T) {
	event.RegisterWriter("mongo", NewMongoWriter)

	conf := &event.EmitterConfig{
		Sender: &event.DriverConfig{
			Type: "logger",
		},
		Writer: &event.DriverConfig{
			Type: "mongo",
			Config: map[string]interface{}{
				"collection": "outbox",
				"connection": map[string]interface{}{
					"uri":      "mongodb://localhost:27017",
					"database": "test",
					"name":     "event",
				},
			},
		},
	}

	ctx := context.Background()

	em, err := event.New(ctx, conf)
	require.Nil(t, err)
	require.NotNil(t, em)

	var buf bytes.Buffer
	logrus.SetOutput(&buf)

	key := fmt.Sprintf("%v", time.Now().Unix())

	require.Nil(t, em.Publish(ctx, "test", key, "testdata", nil))

	time.Sleep(1 * time.Second)
	logStr := buf.String()

	assert.Contains(t, logStr, "key="+key)
	assert.Contains(t, logStr, "data=testdata")
	assert.Contains(t, logStr, "topic=test")

	mconf := client.MongoClient{
		URI:     "mongodb://localhost:27017",
		AppName: "test",
	}
	cl, err := mconf.MongoConnect()
	require.Nil(t, err)

	db := cl.Database(mconf.AppName)
	err = db.Collection("outbox").FindOne(ctx, bson.M{"key": key}).Err()
	require.NotNil(t, err)
	assert.Equal(t, err, mongo.ErrNoDocuments)
	db.Collection("outbox").DeleteMany(ctx, bson.D{})
}
