package validation

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_isValidPhone(t *testing.T) {
	type args struct {
		phone string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "should accept number with + prefix ",
			args: args{
				phone: "+628123123123",
			},
			want: true,
		},
		{
			name: "should accept number length 15 ",
			args: args{
				phone: "878818078668111",
			},
			want: true,
		},
		{
			name: "should accept number length 12 ",
			args: args{
				phone: "878818078668",
			},
			want: true,
		},
		{
			name: "should accept number length 11 ",
			args: args{
				phone: "87881807866",
			},
			want: true,
		},
		{
			name: "should accept number length 10 ",
			args: args{
				phone: "8174050134",
			},
			want: true,
		},
		{
			name: "should accept number length 9 ",
			args: args{
				phone: "817405013",
			},
			want: true,
		},
		{
			name: "should accept number length 7 ",
			args: args{
				phone: "8174050",
			},
			want: true,
		},
		{
			name: "should accept number without + prefix ",
			args: args{
				phone: "8123123123",
			},
			want: true,
		},
		{
			name: "should not accept number with 5 length ",
			args: args{
				phone: "08123",
			},
			want: false,
		},
		{
			name: "should not accept number with 20 length ",
			args: args{
				phone: "08123081230812308123",
			},
			want: false,
		},

		{
			name: "should not accept number with wrong prefix ",
			args: args{
				phone: "*08123123123",
			},
			want: false,
		},
		{
			name: "should not accept number length 16 ",
			args: args{
				phone: "8788180786681112",
			},
			want: false,
		},
		{
			name: "should not accept number length 6 ",
			args: args{
				phone: "878818",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, IsValidPhone(tt.args.phone), "isValidPhone(%v)", tt.args.phone)
		})
	}
}
