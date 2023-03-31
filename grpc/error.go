package grpc

import (
	"context"

	"google.golang.org/grpc"

	"github.com/bondhan/golib/log"

	"github.com/bondhan/golib/errorlib"
)

func NewErrorMappingInterceptor(em map[string]string) grpc.UnaryClientInterceptor {
	logger := log.GetLogger(context.Background(), "grpc", "NewErrorMappingInterceptor")
	logger.Info("return")
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		err := invoker(ctx, method, req, reply, cc, opts...)
		if be, ok := errorlib.DecodeErr(err); ok == nil {
			if code, ok := em[be.ServiceCode]; ok {
				logger.Debugf("mapping error from %s to %s", be.Code, code)
				be.Code = code
				return be.Error()
			}
		}
		return err
	}
}
