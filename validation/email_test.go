package validation

import "testing"

func TestIsValidEmail(t *testing.T) {
	type args struct {
		email        string
		validDomains []string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "valid email in domains",
			args: args{
				email:        "warung@gotoko.co.id",
				validDomains: []string{"gotoko.co.id"},
			},
			want: true,
		},
		{
			name: "valid email with empty domains",
			args: args{
				email:        "warung@gotoko.co.id",
				validDomains: []string{},
			},
			want: true,
		},
		{
			name: "invalid email with symbols",
			args: args{
				email:        "war#ung@gotoko.co.id",
				validDomains: []string{},
			},
			want: false,
		},
		{
			name: "invalid email with double @",
			args: args{
				email:        "war@ung@gotoko.co.id",
				validDomains: []string{"gotoko.co.id"},
			},
			want: false,
		},
		{
			name: "invalid email with double @ but empty domains",
			args: args{
				email:        "war@ung@gotoko.co.id",
				validDomains: []string{},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidEmail(tt.args.email, tt.args.validDomains); got != tt.want {
				t.Errorf("IsValidEmail() = %v, want %v", got, tt.want)
			}
		})
	}
}
