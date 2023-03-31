package client

import (
	"context"
	"log"

	"cloud.google.com/go/storage"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
)

func GStorageClient(ctx context.Context) *storage.Client {
	c, err := storage.NewClient(ctx,
		option.WithGRPCDialOption(grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor())),
		option.WithGRPCDialOption(grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor())),
	)
	if err != nil {
		log.Fatalf("Failed to create bucket client: %v", err)
	}
	return c
}
