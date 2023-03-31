package index

import (
	"context"
	"errors"

	"github.com/bondhan/golib/config"
	"github.com/bondhan/golib/docstore"
)

type DriverFactory func(config interface{}) (Driver, error)

var drivers = map[string]DriverFactory{}

type Driver interface {
	Search(ctx context.Context, text string, query *docstore.QueryOpt) (SearchResult, error)
}

type SearchResult interface {
	Decode(ctx context.Context, docs interface{}) error
	Next(ctx context.Context, doc interface{}) (float64, error)
	Count() int
}

func Register(name string, fn DriverFactory) {
	drivers[name] = fn
}

func GetIndexService(conf config.Getter) (Driver, error) {
	driver := conf.GetString("driver")
	if driver == "" {
		return nil, errors.New("[index] no driver type defined")
	}

	fn, ok := drivers[driver]
	if !ok {
		return nil, errors.New("[index] driver not found")
	}

	return fn(conf)
}
