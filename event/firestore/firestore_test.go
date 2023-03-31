//go:build integration

package firestore

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/iterator"

	"github.com/bondhan/golib/client"
	"github.com/bondhan/golib/event"
)

const emuHost = "localhost:8568"

func TestFireSender(t *testing.T) {
	os.Setenv("FIRESTORE_EMULATOR_HOST", emuHost)
	event.RegisterSender("firestore", NewFireSender)

	fs := client.FirestoreClient(context.Background(), "my-project")
	fsconf := &FireSender{
		Collection: "outbox",
		store:      fs.Collection("outbox"),
	}

	iter := fsconf.store.DocumentRefs(context.Background())
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		doc.Delete(context.Background())

	}

	conf := &event.EmitterConfig{
		Sender: &event.DriverConfig{
			Type:   "firestore",
			Config: fsconf,
		},
	}

	ctx := context.Background()

	em, err := event.New(ctx, conf)
	require.Nil(t, err)
	require.NotNil(t, em)

	key := fmt.Sprintf("%v", time.Now().Unix())

	require.Nil(t, em.Publish(ctx, "test", key, "testdata", nil))

	iter = fsconf.store.DocumentRefs(context.Background())
	count := 0
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		ref, _ := doc.Get(ctx)
		d := ref.Data()
		assert.Equal(t, key, d["key"])
		count++

	}
	assert.Equal(t, 1, count)
	// assert.False(t, true)

}

func TestFireWriter(t *testing.T) {
	os.Setenv("FIRESTORE_EMULATOR_HOST", emuHost)
	event.RegisterWriter("firestore", NewFireWriter)

	fs := client.FirestoreClient(context.Background(), "my-project")
	fsconf := &FireSender{
		Collection: "outbox",
		store:      fs.Collection("outbox"),
	}

	iter := fsconf.store.DocumentRefs(context.Background())
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		doc.Delete(context.Background())

	}

	conf := &event.EmitterConfig{
		Sender: &event.DriverConfig{
			Type: "logger",
		},
		Writer: &event.DriverConfig{
			Type:   "firestore",
			Config: fsconf,
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

	iter = fsconf.store.DocumentRefs(context.Background())
	count := 0
	for {
		_, err := iter.Next()
		if err == iterator.Done {
			break
		}

		count++

	}
	assert.Equal(t, 0, count)
}
