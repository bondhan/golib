package grpc

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"

	"github.com/bondhan/golib/cache"
	"github.com/bondhan/golib/cache/driver"
	"github.com/bondhan/golib/log"
	"github.com/bondhan/golib/util"
)

type ctxKey string

func (c ctxKey) String() string {
	return "gRPC client context " + string(c)
}

var skipCacheCtx = ctxKey("skipCache")

func SkipCache(ctx context.Context) context.Context {
	return context.WithValue(ctx, skipCacheCtx, true)
}

func isSkipCache(ctx context.Context) bool {
	if ctx == nil {
		return false
	}

	if _, ok := ctx.Value(skipCacheCtx).(bool); ok {
		return true
	}

	return false
}

type CachedClient struct {
	cache    *cache.Cache
	duration int
	logger   *logrus.Entry
}

func (c *CachedClient) save(ctx context.Context, method string, req, reply interface{}) {
	var id string

	method = strings.ReplaceAll(strings.TrimPrefix(method, "/"), "/", ".")
	if rq, ok := req.(proto.Message); ok {
		b, err := ProtobufToJSON(rq)
		if err == nil {
			id = util.Hash58(b)
		}
	}

	if rs, ok := reply.(proto.Message); ok && id != "" {
		rep := JsonpbMarshalleble{Message: rs}

		if err := c.cache.Set(ctx, method+":"+id, &rep, c.duration); err != nil {
			c.logger.Warnf("Error storing cache due to: %s", err.Error())

			return
		}
	}
}

func (c *CachedClient) getStoredResponse(ctx context.Context, method string, req, reply interface{}) error {
	if c.cache == nil {
		c.logger.Error("cache is not initialized")
		return errors.New("cache is not initialized")
	}

	method = strings.ReplaceAll(strings.TrimPrefix(method, "/"), "/", ".")

	var id string

	if req != nil {
		if rq, ok := req.(proto.Message); ok {
			b, err := ProtobufToJSON(rq)
			if err == nil {
				id = util.Hash58(b)
			}
		}
	}

	if id == "" {
		c.logger.Error("unable to read request, id is empty")

		return errors.New("unable to read request, id is empty")
	}

	if reply == nil {
		c.logger.Error("reply object is nil")

		return errors.New("reply object is nil")
	}

	if rs, ok := reply.(proto.Message); ok && id != "" {
		rep := JsonpbMarshalleble{Message: rs}

		if err := c.cache.Get(ctx, method+":"+id, &rep); err != nil {
			if !errors.Is(err, driver.NotFound) {
				c.logger.WithError(err).Errorf("Error getting cache %v:%v", method, id)
			}

			return fmt.Errorf("error getting cache %v:%v , %v", method, id, err)
		}

		reply = rep.Message

		if reply == nil {
			c.logger.Error("nil return from cache")

			return fmt.Errorf("nil return from cache %v:%v", method, id)
		}
	}

	return nil
}

func NewCachedClientInterceptor(url string, duration int) grpc.UnaryClientInterceptor {
	logger := log.GetLogger(context.Background(), "grpc", "NewCachedClientInterceptor")

	if url == "" {
		url = "mem://"
	}

	if duration == 0 {
		duration = 60 * 15
	}

	ch, err := cache.New(url)
	if err != nil {
		logger.WithError(err).Error("failed to create new cache")
		return nil
	}

	return NewClientInterceptorWithCache(ch, duration)
}

func NewClientInterceptorWithCache(ch *cache.Cache, duration int) grpc.UnaryClientInterceptor {
	logger := log.GetLogger(context.Background(), "grpc", "NewCachedClientInterceptorWithCache")

	if duration == 0 {
		duration = 60 * 15
	}

	cclient := &CachedClient{cache: ch, duration: duration, logger: logger}

	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		logger.Info("checking cache for ", method, req)

		if cclient != nil && !isSkipCache(ctx) {
			if err := cclient.getStoredResponse(ctx, method, req, reply); err == nil {
				logger.Info("Using cached response")

				return nil
			}
		}

		logger.Info("invoking the grpc method ", method, req)

		err := invoker(ctx, method, req, reply, cc, opts...)
		if err != nil {
			logger.WithError(err).Error("error calling grpc")
		}

		logger.Info("storing result to cache ", method, req)

		if err == nil && cclient != nil {
			cclient.save(ctx, method, req, reply)
		}

		return err
	}
}
