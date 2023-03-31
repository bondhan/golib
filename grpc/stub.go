package grpc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/bondhan/golib/log"
	"github.com/bondhan/golib/util"
)

type JsonpbMarshalleble struct {
	proto.Message
}

func ProtobufToJSON(message proto.Message) ([]byte, error) {
	b, err := protojson.Marshal(message)
	return b, err
}

func (j *JsonpbMarshalleble) MarshalJSON() ([]byte, error) {
	return ProtobufToJSON(j.Message)
}

func (j *JsonpbMarshalleble) UnmarshalJSON(b []byte) error {
	return protojson.Unmarshal(b, j.Message)
}

type Recorder struct {
	storage StubStorage
}

func NewRecorderClientInterceptor(path string) grpc.UnaryClientInterceptor {
	logger := log.GetLogger(context.Background(), "grpc", "NewRecorderClientInterceptor")
	db, err := NewBlobStubStorage(path)
	if err != nil {
		logger.WithError(err).Error("failed to create new stub storage")
		return nil
	}
	rec := &Recorder{
		storage: db,
	}
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		err := invoker(ctx, method, req, reply, cc, opts...)
		rec.recordCall(method, req, reply)
		return err
	}
}

func NewStubClientInterceptor(path string) grpc.UnaryClientInterceptor {
	logger := log.GetLogger(context.Background(), "grpc", "NewStubClientInterceptor")
	db, err := NewBlobStubStorage(path)
	if err != nil {
		logger.WithError(err).Error("failed to create new stub storage")
		return nil
	}
	rec := &Recorder{
		storage: db,
	}
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if err := rec.getStoredResponse(method, req, reply); err == nil {
			return nil
		}

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func NewRecorderServerInterceptor(path string) grpc.UnaryServerInterceptor {

	db, err := NewBlobStubStorage(path)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	rec := &Recorder{
		storage: db,
	}
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		resp, err = handler(ctx, req)
		rec.recordCall(info.FullMethod, req, resp)
		return resp, err
	}
}

func (r *Recorder) getStoredResponse(method string, req, reply interface{}) error {
	method = strings.ReplaceAll(strings.TrimPrefix(method, "/"), "/", ".")
	ms := strings.Split(method, ".")
	svc := strings.Join(ms[:len(ms)-1], ".")
	method = ms[len(ms)-1]

	if rs, ok := reply.(proto.Message); ok {
		rep := JsonpbMarshalleble{Message: rs}
		b, err := r.getResponseStub(svc, method, req)
		if err != nil {
			fmt.Println("[stub] Error getting response ", err)
			return err
		}

		if err := json.Unmarshal(b, &rep); err != nil {
			fmt.Println("[stub] Error unmarshal response ", err)
			return err
		}

		reply = rep.Message
	}
	return nil
}

func (r *Recorder) getResponseStub(svc, method string, req interface{}) ([]byte, error) {
	id := ""
	in := make(map[string]interface{})
	if rq, ok := req.(proto.Message); ok {
		b, err := ProtobufToJSON(rq)
		if err != nil {
			return nil, err
		}
		id = util.Hash58(b)
		if err := json.Unmarshal(b, &in); err != nil {
			return nil, err
		}
	}

	b, err := r.storage.Get(context.Background(), svc, method, "response", id)
	if err != nil {
		rule := r.storage.FindStubs(svc, method, in)
		if rule == nil {
			return nil, errors.New("no matched stub")
		}
		return rule.OutData, nil
	}
	return b, nil
}

func (r *Recorder) recordCall(method string, req, reply interface{}) {
	id := ""
	method = strings.ReplaceAll(strings.TrimPrefix(method, "/"), "/", ".")
	ms := strings.Split(method, ".")
	svc := strings.Join(ms[:len(ms)-1], ".")
	method = ms[len(ms)-1]
	if rq, ok := req.(proto.Message); ok {
		b, err := ProtobufToJSON(rq)
		if err == nil {
			id = util.Hash58(b)
			rqs := JsonpbMarshalleble{Message: rq}
			if b, err := json.MarshalIndent(&rqs, "", "\t"); err == nil {
				if err := r.storage.Write(context.Background(), svc, method, "request", id, b); err != nil {
					fmt.Println("[stub] Error writing request to DB ", err)
				}
			}
		}

	}

	if rs, ok := reply.(proto.Message); ok && id != "" {
		rep := JsonpbMarshalleble{Message: rs}
		if b, err := json.MarshalIndent(&rep, "", "\t"); err == nil {

			if err := r.storage.Write(context.Background(), svc, method, "response", id, b); err != nil {
				fmt.Println("[stub] Error writing response to storage ", err)
			}
		}
	}
}

type StubRule struct {
	In      *Input `json:"in"`
	Out     string `json:"out"`
	OutData []byte
}

type Input struct {
	Equals   map[string]interface{} `json:"equals"`
	Contains map[string]interface{} `json:"contains"`
	Matches  map[string]interface{} `json:"matches"`
}

func (s *StubRule) Match(in map[string]interface{}) bool {

	if s.In == nil {
		return false
	}

	if s.In.Equals != nil {
		return equals(s.In.Equals, in)
	}

	if s.In.Contains != nil {
		return contains(s.In.Contains, in)
	}

	if s.In.Matches != nil {
		return matches(s.In.Matches, in)
	}

	fmt.Println("[stub] no patern to match")

	return false
}

func equals(pattern, in map[string]interface{}) bool {
	if len(pattern) != len(in) {
		return false
	}
	for k, v := range pattern {
		switch val := v.(type) {
		case map[string]interface{}:
			iv, ok := in[k].(map[string]interface{})
			if !ok {
				return false
			}
			if !equals(val, iv) {
				return false
			}
		default:
			if fmt.Sprintf("%v", v) != fmt.Sprintf("%v", in[k]) {
				return false
			}
		}
	}
	return true
}

func contains(pattern, in map[string]interface{}) bool {
	for k, v := range pattern {
		iv, ok := in[k]
		if !ok {
			return false
		}
		switch val := v.(type) {
		case map[string]interface{}:
			if p, ok := iv.(map[string]interface{}); ok {
				return contains(val, p)
			}
			return false
		default:
			if fmt.Sprintf("%v", v) != fmt.Sprintf("%v", iv) {
				return false
			}
		}
	}
	return true
}

func matches(pattern, in map[string]interface{}) bool {
	for k, v := range pattern {
		iv, ok := in[k]
		if !ok {
			return false
		}
		switch val := v.(type) {
		case map[string]interface{}:
			if p, ok := iv.(map[string]interface{}); ok {
				return matches(val, p)
			}
			return false
		case string:
			vStr, ok := iv.(string)
			if !ok {
				return false
			}

			match, err := regexp.Match(val, []byte(vStr))
			if err != nil {
				fmt.Printf("match regexp '%s' with '%s' failed: %v", vStr, val, err)
			}

			return match
		default:
			return false
		}

	}
	return true
}
