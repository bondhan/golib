package redis

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/bondhan/golib/lock"
	redis "github.com/go-redis/redis/v8"
)

const schema = "redis"

// LockManager Redis lock manager
type LockManager struct {
	manager *RedLock
	prefix  string
}

func init() {
	lock.Register(schema, NewDlock)
}

func NewDlock(urls []*url.URL) (lock.DLocker, error) {
	return New(urls)
}

// New create redis locker instance
func New(urls []*url.URL) (*LockManager, error) {
	hs := make([]string, 0)
	for _, u := range urls {
		host := "tcp://"

		pass, ok := u.User.Password()
		if ok {
			host += u.User.Username() + ":" + pass + "@"
		}

		host += u.Host
		hs = append(hs, host)
	}

	lockMgr, err := NewRedLock(hs)
	if err != nil {
		return nil, err
	}

	path := strings.ReplaceAll(urls[0].Path, "/", "")
	if !strings.HasSuffix(path, ":") {
		path += ":"
	}

	return &LockManager{
		manager: lockMgr,
		prefix:  path,
	}, nil
}

// TryLock try to lock, and return immediately if resource already locked
func (l *LockManager) TryLock(ctx context.Context, id string, ttl int) error {
	_, err := l.manager.TryLock(ctx, l.prefix+id, time.Duration(ttl)*time.Second)
	if errors.Is(err, ErrAcquireLock) {
		return lock.ErrResourceLocked
	}
	return err
}

// Lock try to lock and wait until resource is available to lock
func (l *LockManager) Lock(ctx context.Context, id string, ttl int) error {
	_, err := l.manager.Lock(ctx, l.prefix+id, time.Duration(ttl)*time.Second)
	if errors.Is(err, ErrAcquireLock) {
		return lock.ErrResourceLocked
	}
	return err
}

func (l *LockManager) ExtendLock(ctx context.Context, id string, ttl int) error {
	err := l.manager.ExtendLock(ctx, l.prefix+id, time.Duration(ttl)*time.Second)
	if errors.Is(err, ErrExtendLock) {
		return lock.ErrResourceLocked
	}
	return err
}

// Unlock unlock resource
func (l *LockManager) Unlock(ctx context.Context, id string) error {
	return l.manager.UnLock(ctx, l.prefix+id)
}

func (l *LockManager) TryMLock(ctx context.Context, ttl int, ids ...string) error {
	for i, r := range ids {
		ids[i] = l.prefix + r
	}
	_, err := l.manager.TryMLock(ctx, time.Duration(ttl)*time.Second, ids...)
	if errors.Is(err, ErrAcquireLock) {
		return lock.ErrResourceLocked
	}
	return err
}

func (l *LockManager) MLock(ctx context.Context, ttl int, ids ...string) error {
	for i, r := range ids {
		ids[i] = l.prefix + r
	}
	_, err := l.manager.MLock(ctx, time.Duration(ttl)*time.Second, ids...)
	if errors.Is(err, ErrAcquireLock) {
		return lock.ErrResourceLocked
	}
	return err
}

func (l *LockManager) MUnlock(ctx context.Context, ids ...string) error {
	for i, r := range ids {
		ids[i] = l.prefix + r
	}
	return l.manager.MUnLock(ctx, ids...)
}

// Close close the lock
func (l *LockManager) Close() error {
	return l.manager.Close()
}

func (l *LockManager) As(i interface{}) bool {
	cl := l.manager.clients[0]

	p, ok := i.(**redis.Client)
	if !ok {
		return false
	}
	*p = cl.cli
	return true
}
