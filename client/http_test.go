package client

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	URL = "https://httpbin.org/"
)

func TestHttpRequest_Get(t *testing.T) {
	type fields struct {
		Method  string
		URL     string
		Payload interface{}
		Headers map[string]string
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *HttpResponse
		wantErr bool
	}{
		{
			name: "normal",
			fields: fields{
				Method: "GET",
				URL:    URL + "get",
			},
			args: args{
				ctx: context.Background(),
			},
			want: &HttpResponse{
				Status: http.StatusOK,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &HttpRequest{
				Method:  tt.fields.Method,
				URL:     tt.fields.URL,
				Payload: tt.fields.Payload,
				Headers: tt.fields.Headers,
			}
			got, err := h.Get(tt.args.ctx)
			if !tt.wantErr {
				assert.Nil(t, err)
				assert.Equal(t, got.Status, tt.want.Status)
			}
		})
	}
}

func TestHttpRequest_Post(t *testing.T) {
	type fields struct {
		Method  string
		URL     string
		Payload interface{}
		Headers map[string]string
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *HttpResponse
		wantErr bool
	}{
		{
			name: "normal",
			fields: fields{
				Method: "POST",
				URL:    URL + "post",
				Payload: map[string]interface{}{
					"name":  "ahmad",
					"age":   10,
					"value": 1.23,
				},
			},
			args: args{
				ctx: context.Background(),
			},
			want: &HttpResponse{
				Status: http.StatusOK,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &HttpRequest{
				Method:  tt.fields.Method,
				URL:     tt.fields.URL,
				Payload: tt.fields.Payload,
				Headers: tt.fields.Headers,
			}
			got, err := h.Post(tt.args.ctx)
			if !tt.wantErr {
				assert.Nil(t, err)
				assert.Equal(t, got.Status, tt.want.Status)
			}
		})
	}
}
