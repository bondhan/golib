package grpc

import (
	"context"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/bondhan/golib/domain/retailer"
)

func getRetailer() *retailer.RetailerContext {
	data := map[string]interface{}{
		"id":     "1234566",
		"number": "RT0001",
		"user":   "user1",
		"address": map[string]interface{}{
			"province":     "Jawa Barat",
			"city":         "Bandung",
			"district":     "Cibeunying Kaler",
			"district_sub": "Cibeunying Kidul",
			"zip_code":     "40132",
		},

		"area": map[string]interface{}{
			"region":     "Jawa Barat",
			"region_sub": "Bandung",
			"cluster":    "Kota",
			"warehouse":  "Dago",
		},
		"segments": []string{"small", "medium"},
	}

	ret, _ := retailer.DecodeRetailer(data)
	return ret

}

func DummyHandler() grpc.UnaryHandler {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		return retailer.GetRetailerFromContext(ctx)
	}
}

type dummyInvokerWrapper struct {
	req interface{}
}

func DummyInvoker(wrapper *dummyInvokerWrapper) grpc.UnaryInvoker {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		wrapper.req = req
		return nil
	}
}

func TestRetailerMiddleware(t *testing.T) {
	ret := getRetailer()
	str, _ := ret.Encode()

	req := &HelloRequest{
		Data: &HelloRequest_Data{
			Headers: map[string]string{retailer.RetailerHeaderKey: str},
			Params: &HelloRequest_Data_Params{
				Id: "1234",
			},
		},
	}
	mw := NewRetailerExtractor("")
	resp, err := mw(context.Background(), req, &grpc.UnaryServerInfo{}, DummyHandler())
	if err != nil {
		t.Error(err)
	}

	if resp == nil {
		t.Error("response should not nil")
	}

	ret, ok := resp.(*retailer.RetailerContext)
	if !ok {
		t.Error("invalid type")
	}
	if ret.ID == "" {
		t.Error("id should not empty")
	}
}

func TestRetailerClientMiddlewre(t *testing.T) {
	clientConn, err := grpc.Dial("fake:connection", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("failed to create client connection: %v", err)
	}

	mw := NewRetailerPropagator("")
	ret := getRetailer()
	ctx := retailer.WithRetailerContext(context.Background(), ret)
	wrapper := &dummyInvokerWrapper{}
	invoker := DummyInvoker(wrapper)

	req := &HelloRequest{
		Data: &HelloRequest_Data{
			Params: &HelloRequest_Data_Params{
				Id: "1234",
			},
		},
	}

	err = mw(ctx, "fake:method", req, &HelloResponse{}, clientConn, invoker)
	if err != nil {
		t.Error(err)
	}

	if wrapper.req == nil {
		t.Error("request should not nil")
	}

	hreq, ok := wrapper.req.(*HelloRequest)
	if !ok {
		t.Error("invalid type")
	}

	if hreq.Data == nil {
		t.Error("data should not nil")
	}

	if hreq.Data.Headers == nil {
		t.Error("headers should not nil")
	}

	retstr := hreq.Data.Headers[retailer.RetailerHeaderKey]
	if retstr == "" {
		t.Error("retailer should not empty")
	}

	enc, _ := ret.Encode()
	if retstr != enc {
		t.Error("retailer should equal")
	}
}
