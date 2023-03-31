package lock

import (
	"context"
	"errors"
	"net/url"
	"strings"
)

var (
	ErrResourceLocked = errors.New("resource locked")
)

// DLocker distributed locker interface
type DLocker interface {
	TryLock(ctx context.Context, id string, ttl int) error
	Lock(ctx context.Context, id string, ttl int) error
	Unlock(ctx context.Context, id string) error
	ExtendLock(ctx context.Context, id string, ttl int) error
	Close() error
	As(i interface{}) bool
}

// InitFunc cache init function
type InitFunc func(urls []*url.URL) (DLocker, error)

var lockerImpl = make(map[string]InitFunc)

// Register register cache implementation
func Register(schema string, f InitFunc) {
	lockerImpl[schema] = f
}

type Locker struct {
	dlocker DLocker
}

// New create new cache
func New(urlStr string) (*Locker, error) {
	if urlStr == "" {
		urlStr = "local://"
	}

	if !strings.HasSuffix(urlStr, "/") {
		urlStr += "/"
	}

	first, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	up := []*url.URL{first}

	if first.Scheme == "local" {
		return &Locker{dlocker: DLocal()}, nil
	}

	urls := strings.Split(strings.TrimPrefix(urlStr, first.Scheme+"://"), ",")

	if len(urls) > 1 {
		scheme := first.Scheme + "://"
		path := first.Path
		up = make([]*url.URL, 0)

		for _, u := range urls {
			if !strings.HasPrefix(u, scheme) {
				u = scheme + u
			}
			if !strings.HasSuffix(u, path) {
				u += path
			}
			ur, err := url.Parse(u)
			if err != nil {
				return nil, err
			}
			if first.User != nil {
				*ur.User = *first.User
			}
			up = append(up, ur)
		}
	}

	f, ok := lockerImpl[first.Scheme]
	if !ok {
		return nil, errors.New("unsupported scheme")
	}

	dl, err := f(up)
	if err != nil {
		return nil, err
	}
	return &Locker{dlocker: dl}, nil
}

func (l *Locker) TryLock(ctx context.Context, id string, ttl int) error {
	return l.dlocker.TryLock(ctx, id, ttl)
}

func (l *Locker) Lock(ctx context.Context, id string, ttl int) error {
	return l.dlocker.Lock(ctx, id, ttl)
}

func (l *Locker) Unlock(ctx context.Context, id string) error {
	return l.dlocker.Unlock(ctx, id)
}

func (l *Locker) ExtendLock(ctx context.Context, id string, ttl int) error {
	return l.dlocker.ExtendLock(ctx, id, ttl)
}

func (l *Locker) Close() error {
	return l.dlocker.Close()
}

func (l *Locker) As(i interface{}) bool {
	return l.dlocker.As(i)
}
