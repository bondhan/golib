package cache

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bondhan/golib/cache/driver"
	"github.com/bondhan/golib/cache/embed"
	"github.com/bondhan/golib/cache/lru"
	"github.com/bondhan/golib/cache/mem"
	"github.com/bondhan/golib/cache/redis"
)

type sleepFunc func(t time.Duration)

func TestMemCache(t *testing.T) {
	url := "mem://"
	c, err := New(url)
	require.Nil(t, err)
	assert.NotNil(t, c)

	mc, ok := c.driver.(*mem.MemoryCache)
	assert.True(t, ok)
	assert.NotNil(t, mc)

	testCache(t, c.driver, func(d time.Duration) {
		time.Sleep(d)
	})
}

func TestRedisCache(t *testing.T) {
	s, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	defer s.Close()

	url := "redis://" + s.Addr()

	c, err := New(url)
	require.Nil(t, err)
	assert.NotNil(t, c)

	rc, ok := c.driver.(*redis.Cache)
	assert.True(t, ok)
	assert.NotNil(t, rc)

	testCache(t, c.driver, func(t time.Duration) {
		s.FastForward(t)
	})
}

func TestLRUCache(t *testing.T) {
	url := "lru://"
	c, err := New(url)
	require.Nil(t, err)
	assert.NotNil(t, c)

	mc, ok := c.driver.(*lru.Cache)
	assert.True(t, ok)
	assert.NotNil(t, mc)

	testCache(t, c.driver, func(t time.Duration) {
		time.Sleep(t)
	})
}

func TestEmbedCache(t *testing.T) {
	url := "embed://mem"
	c, err := New(url)
	require.Nil(t, err)
	assert.NotNil(t, c)

	mc, ok := c.driver.(*embed.BadgerCache)
	assert.True(t, ok)
	assert.NotNil(t, mc)

	testCache(t, c.driver, func(t time.Duration) {
		time.Sleep(t)
	})
}

func testCache(t *testing.T, c driver.CacheDriver, sleep sleepFunc) {
	ctx := context.Background()
	err := c.Set(ctx, "tesstring", "value", 0)
	require.Nil(t, err)
	rstring, err := c.GetString(ctx, "tesstring")
	require.Nil(t, err)
	assert.Equal(t, "value", rstring)
	rbyte, err := c.Get(ctx, "tesstring")
	require.Nil(t, err)
	assert.Equal(t, []byte("value"), rbyte)

	err = c.Set(ctx, "testint", 123, 0)
	require.Nil(t, err)
	rint, err := c.GetInt(ctx, "testint")
	require.Nil(t, err)
	assert.Equal(t, int64(123), rint)

	b, err := c.Get(ctx, "testint")
	require.Nil(t, err)
	assert.Equal(t, "123", string(b))

	err = c.Set(ctx, "testfloat", 10.5, 0)
	require.Nil(t, err)
	rfloat, err := c.GetFloat(ctx, "testfloat")
	require.Nil(t, err)
	assert.Equal(t, 10.5, rfloat)

	b, err = c.Get(ctx, "testfloat")
	require.Nil(t, err)
	assert.Equal(t, "10.5", string(b))

	err = c.Set(ctx, "testbool", true, 0)
	require.Nil(t, err)

	b, err = c.Get(ctx, "testbool")
	require.Nil(t, err)
	assert.Equal(t, "1", string(b))

	assert.True(t, c.Exist(ctx, "tesstring"))
	assert.True(t, c.Exist(ctx, "testint"))

	err = c.Set(ctx, "testexp", "any", 10)
	require.Nil(t, err)
	remain := c.RemainingTime(ctx, "testexp")

	assert.Equal(t, 10, remain)

	sleep(time.Duration(1) * time.Second)

	remain = c.RemainingTime(ctx, "testexp")
	assert.Equal(t, 9, remain)

	obj := map[string]interface{}{
		"env":     "dev",
		"port":    "8080",
		"host":    "localhost",
		"counter": 1,
	}

	err = c.Set(ctx, "testobj", obj, 0)
	require.Nil(t, err)

	var res map[string]interface{}

	err = c.GetObject(ctx, "testobj", &res)
	require.Nil(t, err)

	assert.Equal(t, obj["env"], res["env"])
	assert.Equal(t, obj["port"], res["port"])
	assert.Equal(t, fmt.Sprintf("%v", obj["counter"]), fmt.Sprintf("%v", res["counter"]))
}
