package redis

import (
	"context"
	"errors"
	"math/rand"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	mlockScript = `
		if redis.pcall('EXISTS', unpack(KEYS)) > 0 
		then
			return 0
		end

		for index, value in ipairs(KEYS) do 
			if not redis.pcall('SET', KEYS[index], ARGV[1], 'EX', ARGV[2], 'NX')
			then
				return 0
			end
		end
		return 1
		`
	munlockScript = `
		for index, value in ipairs(KEYS) do 
			local v = redis.pcall('GET', KEYS[index])
			if v == ARGV[1]
			then
				redis.call('DEL', KEYS[index])
			else
				return 0
			end
		end
		return 1
		`
)

func mlockInstance(ctx context.Context, client *RedClient, val string, ttl time.Duration, resources ...string) (bool, error) {
	if len(resources) == 0 {
		return false, errors.New("empty resources")
	}

	script := mlockScript
	dur := formatSec(ctx, ttl)
	if usePrecise(ttl) {
		script = strings.ReplaceAll(mlockScript, "'EX'", "'PX'")
		dur = formatMs(ctx, ttl)
	}

	reply := client.cli.Eval(ctx, script, resources, val, dur)
	if reply.Err() != nil {
		return false, reply.Err()
	}
	res, err := reply.Int()
	if err != nil {
		return false, err
	}

	if res != 1 {
		return false, ErrLockSingleRedis
	}
	return true, nil
}

func munlockInstance(ctx context.Context, client *RedClient, val string, resources ...string) (bool, error) {
	if len(resources) == 0 {
		return false, errors.New("empty resources")
	}
	reply := client.cli.Eval(ctx, munlockScript, resources, val)
	if reply.Err() != nil {
		return false, reply.Err()
	}
	return true, nil
}

func (r *RedLock) tryMLock(ctx context.Context, val string, ttl time.Duration, resources ...string) (int64, error) {
	start := time.Now()
	ctxCancel := int32(0)
	success := int32(0)
	cctx, cancel := context.WithTimeout(ctx, ttl)
	var wg sync.WaitGroup
	for _, cli := range r.clients {
		cli := cli
		wg.Add(1)
		go func() {
			defer wg.Done()
			locked, err := mlockInstance(cctx, cli, val, ttl, resources...) // nolint:errcheck
			if err == context.Canceled {
				atomic.AddInt32(&ctxCancel, 1)
			}
			if locked {
				atomic.AddInt32(&success, 1)
			}
		}()
	}
	wg.Wait()
	cancel()
	// fast fail, terminate acquiring lock if context is canceled
	if atomic.LoadInt32(&ctxCancel) > int32(0) {
		return 0, context.Canceled
	}

	drift := int(float64(ttl)*r.driftFactor) + 2
	costTime := time.Since(start).Nanoseconds()
	validityTime := int64(ttl) - costTime - int64(drift)
	if int(success) >= r.quorum && validityTime > 0 {
		r.cache.Set(strings.Join(resources, ":"), val, validityTime)
		return validityTime, nil
	}
	cctx, cancel = context.WithTimeout(ctx, ttl)
	for _, cli := range r.clients {
		cli := cli
		wg.Add(1)
		go func() {
			defer wg.Done()
			munlockInstance(cctx, cli, val, resources...) // nolint:errcheck
		}()
	}
	wg.Wait()
	cancel()
	return 0, ErrAcquireLock
}

// TryLock try to acquire lock
func (r *RedLock) TryMLock(ctx context.Context, ttl time.Duration, resources ...string) (int64, error) {
	val := getRandStr()
	return r.tryMLock(ctx, val, ttl, resources...)
}

// Lock acquires a distribute lock
func (r *RedLock) MLock(ctx context.Context, ttl time.Duration, resources ...string) (int64, error) {
	val := getRandStr()

	rcount := int(ttl) / r.retryDelay

	for i := 0; i < rcount; i++ {
		v, err := r.tryMLock(ctx, val, ttl, resources...)
		if err == nil {
			return v, nil
		}
		// Wait a random delay before to retry
		time.Sleep(time.Duration(rand.Intn(r.retryDelay)) * time.Millisecond)
	}

	return 0, ErrAcquireLock
}

// UnLock releases an acquired lock
func (r *RedLock) MUnLock(ctx context.Context, resources ...string) error {
	rid := strings.Join(resources, ":")
	elem, err := r.cache.Get(rid)
	if err != nil {
		return err
	}
	if elem == nil {
		return nil
	}
	defer r.cache.Delete(rid)

	var wg sync.WaitGroup
	for _, cli := range r.clients {
		cli := cli
		wg.Add(1)
		go func() {
			defer wg.Done()
			munlockInstance(ctx, cli, elem.Val, resources...) //nolint:errcheck
		}()
	}
	wg.Wait()

	return nil
}
