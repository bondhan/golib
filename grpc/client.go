package grpc

import (
	"log"

	mw "github.com/grpc-ecosystem/go-grpc-middleware"
	retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type clientOption struct {
	userAgent     string
	recordPath    string
	stubPath      string
	cacheURL      string
	cacheDuration int
	retryMax      uint
	middlewares   []grpc.UnaryClientInterceptor
	errorMapping  map[string]string
}

type ClientOpt func(*clientOption)

func WithUserAgent(ua string) ClientOpt {
	return func(co *clientOption) {
		co.userAgent = ua
	}
}

func WithClientRecorder(path string) ClientOpt {
	return func(co *clientOption) {
		co.recordPath = path
	}
}

func WithClientStub(path string) ClientOpt {
	return func(co *clientOption) {
		co.stubPath = path
	}
}

func WithClientCache(cacheUrl string, duration int) ClientOpt {
	return func(co *clientOption) {
		co.cacheURL = cacheUrl
		if duration > 0 {
			co.cacheDuration = duration
		}
	}
}

func WithClientMiddleware(mw ...grpc.UnaryClientInterceptor) ClientOpt {
	return func(co *clientOption) {
		co.middlewares = mw
	}
}

func WithErrorMapping(em map[string]string) ClientOpt {
	return func(co *clientOption) {
		co.errorMapping = em
	}
}

func ClientDial(target string, opts ...ClientOpt) *grpc.ClientConn {
	opt := &clientOption{retryMax: 3, cacheDuration: 60 * 5}
	for _, o := range opts {
		o(opt)
	}

	uchain := []grpc.UnaryClientInterceptor{
		otelgrpc.UnaryClientInterceptor(),
	}

	if opt.cacheURL != "" {
		cachedClientInterceptor := NewCachedClientInterceptor(opt.cacheURL, opt.cacheDuration)
		if cachedClientInterceptor != nil {
			uchain = append(uchain, cachedClientInterceptor)
		}
	}

	if opt.retryMax > 0 {
		uchain = append(uchain, retry.UnaryClientInterceptor(retry.WithMax(opt.retryMax)))
	}

	if opt.recordPath != "" {
		recorderClientInterceptor := NewRecorderClientInterceptor(opt.recordPath)
		if recorderClientInterceptor != nil {
			uchain = append(uchain, recorderClientInterceptor)
		}
	}

	if opt.stubPath != "" {
		stubClientInterceptor := NewStubClientInterceptor(opt.stubPath)
		if stubClientInterceptor != nil {
			uchain = append(uchain, stubClientInterceptor)
		}
	}

	if len(opt.middlewares) > 0 {
		uchain = append(uchain, opt.middlewares...)
	}

	if len(opt.errorMapping) > 0 {
		errorMappingInterceptor := NewErrorMappingInterceptor(opt.errorMapping)
		if errorMappingInterceptor != nil {
			uchain = append(uchain, errorMappingInterceptor)
		}
	}

	conn, err := grpc.Dial(target, grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUserAgent(opt.userAgent),
		grpc.WithUnaryInterceptor(
			mw.ChainUnaryClient(uchain...)),
		grpc.WithStreamInterceptor(
			mw.ChainStreamClient(
				otelgrpc.StreamClientInterceptor(),
				retry.StreamClientInterceptor(retry.WithMax(opt.retryMax)),
			)),
	)

	if err != nil {
		log.Fatalf("failed connecting to %s : %v", target, err)
	}
	return conn
}
