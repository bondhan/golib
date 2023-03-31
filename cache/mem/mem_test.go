package mem

import (
	"fmt"
	"math/rand"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCacheURL(t *testing.T) {
	urlStr := "mem://"
	u, err := url.Parse(urlStr)
	assert.Nil(t, err)
	assert.NotNil(t, u)

	cache, err := NewCache(u)
	assert.Nil(t, err)
	assert.NotNil(t, cache)

	mc, ok := cache.(*MemoryCache)
	assert.True(t, ok)
	assert.NotNil(t, mc)
}

func TestMemoryCacheGetSet(t *testing.T) {
	dCache := NewMemoryCache()
	tests := []struct {
		key   string
		value interface{}
		time  int
	}{
		{
			key:   "test.string",
			value: "value",
			time:  1,
		},
		{
			key:   "test.integer",
			value: int64(1),
			time:  1,
		},
		{
			key:   "test.float",
			value: float64(1.1),
			time:  1,
		},
	}

	for _, tt := range tests {
		err := dCache.Set(nil, tt.key, tt.value, tt.time)
		assert.Nil(t, err)

		switch tt.value.(type) {
		case string:
			rstring, err := dCache.GetString(nil, tt.key)
			assert.Nil(t, err)
			assert.Equal(t, tt.value, rstring)
		case int64:
			rint, err := dCache.GetInt(nil, tt.key)
			assert.Nil(t, err)
			assert.Equal(t, tt.value.(int64), rint)
		case float64:
			rfloat, err := dCache.GetFloat(nil, tt.key)
			assert.Nil(t, err)
			assert.Equal(t, tt.value.(float64), rfloat)
		}

		assert.True(t, dCache.Exist(nil, tt.key))
	}

	err := dCache.Set(nil, tests[0].key, tests[0].value, tests[0].time)
	require.NoError(t, err)
}

func TestMemoryCacheExpiry(t *testing.T) {
	dCache := NewMemoryCache()
	tests := []struct {
		key   string
		value interface{}
		time  int
	}{
		{
			key:   "test.exp",
			value: "foo",
			time:  3,
		},
	}

	for _, tt := range tests {
		err := dCache.Set(nil, tt.key, tt.value, tt.time)
		assert.Nil(t, err)

		remain := dCache.RemainingTime(nil, tt.key)
		assert.Equal(t, tt.time, remain)

		time.Sleep(time.Duration(1) * time.Second)
		remain = dCache.RemainingTime(nil, tt.key)
		assert.Equal(t, tt.time-1, remain)

		time.Sleep(time.Duration(3) * time.Second)
		remain = dCache.RemainingTime(nil, tt.key)
		assert.Equal(t, -1, remain)
		assert.Equal(t, dCache.Exist(nil, tt.key), false)
	}
}

func TestMemoryCacheDelete(t *testing.T) {
	dCache := NewMemoryCache()
	tests := []struct {
		key   string
		value interface{}
		time  int
	}{
		{
			key:   "test.delete",
			value: "foo",
			time:  3,
		},
	}

	for _, tt := range tests {
		err := dCache.Set(nil, tt.key, tt.value, tt.time)
		assert.Nil(t, err)

		err = dCache.Delete(nil, tt.key)
		assert.Nil(t, err)
		assert.Equal(t, dCache.Exist(nil, tt.key), false)
	}
}

func TestMemObject(t *testing.T) {
	dCache := NewMemoryCache()
	assert.NotNil(t, dCache)

	obj := map[string]interface{}{
		"env":     "dev",
		"port":    "8080",
		"host":    "localhost",
		"counter": 1,
	}

	err := dCache.Set(nil, "testobj", obj, 0)
	assert.Nil(t, err)

	var res map[string]interface{}

	err = dCache.GetObject(nil, "testobj", &res)
	assert.Nil(t, err)

	assert.Equal(t, obj["env"], res["env"])
	assert.Equal(t, obj["port"], res["port"])
	assert.Equal(t, fmt.Sprintf("%v", obj["counter"]), fmt.Sprintf("%v", res["counter"]))
}

func TestMemBytes(t *testing.T) {
	dCache := NewMemoryCache()
	assert.NotNil(t, dCache)

	err := dCache.Set(nil, "testint", 10, 0)
	assert.Nil(t, err)

	b, err := dCache.Get(nil, "testint")
	assert.Nil(t, err)
	assert.Equal(t, "10", string(b))

	err = dCache.Set(nil, "testbool", true, 0)
	assert.Nil(t, err)

	b, err = dCache.Get(nil, "testbool")
	assert.Nil(t, err)
	assert.Equal(t, "1", string(b))
}

type cacheSeed struct {
	key  string
	val  interface{}
	time int
}

func TestLongLivedCache(t *testing.T) {
	dCache := NewMemoryCache()

	n := 1000 * 100
	data := make([]cacheSeed, 0, n)

	for i := 0; i < n; i++ {
		data = append(data, cacheSeed{
			key:  fmt.Sprintf("%d", i),
			val:  i,
			time: rand.Intn(3) + 1,
		})
	}

	for _, d := range data {
		err := dCache.Set(nil, d.key, d.val, d.time)
		assert.NoError(t, err)
	}

	time.Sleep(5 * time.Second)
	assert.Equal(t, len(dCache.data), 0)
}
