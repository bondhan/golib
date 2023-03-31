package grpc

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"gocloud.dev/blob"
)

type StubStorage interface {
	Get(ctx context.Context, service, method, kind, id string) ([]byte, error)
	Write(ctx context.Context, service, method, kind, id string, data []byte) error
	FindStubs(service, method string, in map[string]interface{}) *StubRule
}

type BlobStubStorage struct {
	bucket *blob.Bucket
	stubs  map[string][]*StubRule
}

func NewBlobStubStorage(conURL string) (StubStorage, error) {

	bucket, err := blob.OpenBucket(context.Background(), conURL)
	if err != nil {
		return nil, err
	}

	store := &BlobStubStorage{
		bucket: bucket,
		stubs:  make(map[string][]*StubRule, 0),
	}

	if err := store.LoadStubs(context.Background()); err != nil {
		return nil, err
	}

	return store, nil
}

func (b *BlobStubStorage) Get(ctx context.Context, service, method, kind, id string) ([]byte, error) {
	buck := blob.PrefixedBucket(b.bucket, service+"/"+method+"/"+kind+"/")
	return buck.ReadAll(ctx, id+".json")
}

func (b *BlobStubStorage) Write(ctx context.Context, service, method, kind, id string, data []byte) error {
	buck := blob.PrefixedBucket(b.bucket, service+"/"+method+"/"+kind+"/")
	return buck.WriteAll(ctx, id+".json", data, nil)
}

func (b *BlobStubStorage) LoadStubs(ctx context.Context) error {
	buck := blob.PrefixedBucket(b.bucket, "rules/")
	iter := buck.List(nil)

	for {
		obj, err := iter.Next(ctx)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		rb, err := buck.ReadAll(ctx, obj.Key)
		if err != nil {
			fmt.Println("[stub] error read rule file", obj.Key)
			continue
		}
		srule := make([]StubRule, 0)
		if err := json.Unmarshal(rb, &srule); err != nil {
			fmt.Println("[stub] error unmarshal rule", obj.Key)
			continue
		}

		id := strings.TrimSuffix(obj.Key, ".json")
		ms := strings.Split(id, ".")
		svc := strings.Join(ms[:len(ms)-1], ".")
		method := ms[len(ms)-1]

		for _, r := range srule {
			if r.Out == "" {
				continue
			}

			rbuck := blob.PrefixedBucket(b.bucket, svc+"/"+method+"/response/")
			ob, err := rbuck.ReadAll(ctx, r.Out+".json")
			if err != nil {
				fmt.Println("[stub] error read response data ", r.Out)
				continue
			}
			r.OutData = ob
			if b.stubs[id] == nil {
				b.stubs[id] = make([]*StubRule, 0)
			}
			b.stubs[id] = append(b.stubs[id], &r)
		}

	}
	return nil
}

func (b *BlobStubStorage) FindStubs(service, method string, in map[string]interface{}) *StubRule {
	stubs, ok := b.stubs[service+"."+method]
	if !ok {
		return nil
	}

	if len(in) == 0 {
		return stubs[0]
	}

	for _, stub := range stubs {
		if stub.Match(in) {
			return stub
		}
	}

	return nil
}
