package config

import (
	"reflect"
	"testing"
)

var DefaultConfig = map[string]interface{}{
	"name": map[string]string{
		"testing": "value",
	},
}

func TestEmbedConfig_GetStringMapString(t *testing.T) {
	type fields struct {
		config interface{}
	}
	type args struct {
		k string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   map[string]string
	}{
		{
			name: "success",
			fields: fields{
				config: DefaultConfig,
			},
			args: args{
				k: "name",
			},
			want: map[string]string{
				"testing": "value",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &EmbedConfig{
				config: tt.fields.config,
			}
			if got := e.GetStringMapString(tt.args.k); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("EmbedConfig.GetStringMapString() = %v, want %v", got, tt.want)
			}
		})
	}
}
