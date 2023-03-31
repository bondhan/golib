//go:build integration

package redis

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBasicMLock(t *testing.T) {
	t.Skip()
	ctx := context.Background()
	lock, err := NewRedLock(redisServers)
	require.Nil(t, err)

	_, err = lock.MLock(ctx, 200*time.Millisecond, "foo", "bar")
	assert.Nil(t, err)
	err = lock.MUnLock(ctx, "foo", "bar")
	assert.Nil(t, err)
}

func TestMUnlockExpiredKey(t *testing.T) {
	t.Skip()
	ctx := context.Background()
	lock, err := NewRedLock(redisServers)
	assert.Nil(t, err)

	_, err = lock.MLock(ctx, 50*time.Millisecond, "foo", "bar")
	assert.Nil(t, err)
	time.Sleep(51 * time.Millisecond)
	err = lock.MUnLock(ctx, "foo", "bar")
	assert.Nil(t, err)
}

func TestTryMLock(t *testing.T) {
	t.Skip()
	ctx := context.Background()
	lock, err := NewRedLock(redisServers)
	assert.Nil(t, err)
	_, err = lock.MLock(ctx, 1*time.Second, "foo", "bar")
	assert.Nil(t, err)

	_, err = lock.TryMLock(ctx, 1*time.Second, "bar", "two")
	assert.NotNil(t, err)

	err = lock.MUnLock(ctx, "foo", "bar")
	assert.Nil(t, err)

	_, err = lock.MLock(ctx, 1*time.Second, "bar", "two")
	assert.Nil(t, err)
}
