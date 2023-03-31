package lock

import (
	"context"
	"errors"
	"net/url"
	"strings"
)

type MLocker interface {
	TryMLock(ctx context.Context, ttl int, ids ...string) error
	MLock(ctx context.Context, ttl int, ids ...string) error
	MUnlock(ctx context.Context, ids ...string) error
	Close() error
	As(i interface{}) bool
}

// InitFunc cache init function
type InitFuncM func(urls []*url.URL) (MLocker, error)

var mlockerImpl = make(map[string]InitFuncM)

// Register register cache implementation
func MRegister(schema string, f InitFunc) {
	lockerImpl[schema] = f
}

type MLock struct {
	mlocker MLocker
}

// New create new cache
func NewMLock(urlStr string) (*MLock, error) {
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
		return &MLock{mlocker: Local()}, nil
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

	f, ok := mlockerImpl[first.Scheme]
	if !ok {
		return nil, errors.New("unsupported scheme")
	}

	dl, err := f(up)
	if err != nil {
		return nil, err
	}
	return &MLock{mlocker: dl}, nil
}

func (l *MLock) TryLock(ctx context.Context, ttl int, ids ...string) error {
	return l.mlocker.TryMLock(ctx, ttl, ids...)
}

func (l *MLock) Lock(ctx context.Context, ttl int, ids ...string) error {
	return l.mlocker.MLock(ctx, ttl, ids...)
}

func (l *MLock) Unlock(ctx context.Context, ids ...string) error {
	return l.mlocker.MUnlock(ctx, ids...)
}

func (l *MLock) Close() error {
	return l.mlocker.Close()
}

func (l *MLock) As(i interface{}) bool {
	return l.mlocker.As(i)
}
