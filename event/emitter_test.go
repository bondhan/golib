package event

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDirectEmitter(t *testing.T) {
	conf := &EmitterConfig{
		Sender: &DriverConfig{
			Type: "logger",
		},
	}
	ctx := context.Background()

	em, err := New(ctx, conf)
	require.Nil(t, err)
	require.NotNil(t, em)

	var buf bytes.Buffer
	logrus.SetOutput(&buf)

	require.Nil(t, em.Publish(ctx, "test", "t123", "testdata", nil))
	logStr := buf.String()

	assert.Contains(t, logStr, "key=t123")
	assert.Contains(t, logStr, "data=testdata")
	assert.Contains(t, logStr, "topic=test")
	fmt.Println(logStr)
	//assert.False(t, true)

}

func TestHybridEmitter(t *testing.T) {
	conf := &EmitterConfig{
		Sender: &DriverConfig{
			Type: "logger",
		},
		Writer: &DriverConfig{
			Type: "logger",
		},
	}
	ctx := context.Background()

	em, err := New(ctx, conf)
	require.Nil(t, err)
	require.NotNil(t, em)

	var buf bytes.Buffer
	logrus.SetOutput(&buf)

	require.Nil(t, em.Publish(ctx, "test", "t123", "testdata", nil))
	time.Sleep(1 * time.Second)
	logStr := buf.String()

	assert.Contains(t, logStr, "key=t123")
	assert.Contains(t, logStr, "data=testdata")
	assert.Contains(t, logStr, "topic=test")
	assert.Contains(t, logStr, "message succesfully sent")

}

func TestMetadata(t *testing.T) {
	conf := &EmitterConfig{
		Sender: &DriverConfig{
			Type: "logger",
		},
		EventConfig: &EventConfig{
			Metadata: map[string]map[string]interface{}{
				MetaDefault: {
					"foo": "bar",
				},
			},
		},
	}
	ctx := context.Background()

	em, err := New(ctx, conf)
	require.Nil(t, err)
	require.NotNil(t, em)

	var buf bytes.Buffer
	logrus.SetOutput(&buf)

	require.Nil(t, em.Publish(ctx, "test", "t123", "testdata", map[string]interface{}{
		"baz": "qux",
	}))
	logStr := buf.String()

	assert.Contains(t, logStr, "topic=test")
	assert.Contains(t, logStr, "foo:bar")
	assert.Contains(t, logStr, "baz:qux")

}
