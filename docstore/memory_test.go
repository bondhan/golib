package docstore

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMemoryStore(t *testing.T) {
	ms := NewMemoryStore("test", "id")
	DriverCRUDTest(ms, t)
	DriverBulkTest(ms, t)
}

func TestMemoryStore_Distinct(t *testing.T) {
	type fields struct {
		storage map[interface{}]map[string]interface{}
		idField string
		mux     *sync.Mutex
	}
	type args struct {
		ctx       context.Context
		fieldName string
		filter    interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []interface{}
		wantErr error
	}{
		{
			name:    "assert not error test",
			wantErr: errors.New("[docstore/memory] not implement distinct"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MemoryStore{
				storage: tt.fields.storage,
				idField: tt.fields.idField,
				mux:     tt.fields.mux,
			}
			got, err := m.Distinct(tt.args.ctx, tt.args.fieldName, tt.args.filter)
			if err != nil {
				assert.Equal(t, tt.wantErr, err)
				return
			}
			assert.Equalf(t, tt.want, got, "Distinct(%v, %v, %v)", tt.args.ctx, tt.args.fieldName, tt.args.filter)
		})
	}
}
