package grpc

import (
	"context"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/bondhan/golib/domain/retailer"
)

const defaultPath = "data.headers." + retailer.RetailerHeaderKey

func NewRetailerExtractor(path string) grpc.UnaryServerInterceptor {
	if path == "" {
		path = defaultPath
	}
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		if rq, ok := req.(proto.Message); ok {
			b, err := ProtobufToJSON(rq)
			if err == nil {
				res := gjson.GetBytes(b, path)
				if res.Exists() {
					ctx = retailer.WithRetailerStringContext(ctx, res.String())
				}
			}
		}
		resp, err = handler(ctx, req)

		return resp, err
	}
}

func NewRetailerPropagator(path string) grpc.UnaryClientInterceptor {
	if path == "" {
		path = defaultPath
	}
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if rq, ok := req.(proto.Message); ok {
			b, err := ProtobufToJSON(rq)
			if err == nil {
				ret, err := retailer.GetRetailerFromContext(ctx)
				if err == nil && ret != nil {
					if rstr, err := ret.Encode(); err == nil {
						if out, err := sjson.SetBytes(b, path, rstr); err == nil {
							protojson.Unmarshal(out, rq)
						}
					}
				}
			}
		}
		err := invoker(ctx, method, req, reply, cc, opts...)
		return err
	}
}
