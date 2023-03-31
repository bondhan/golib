package util

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewRegularFile(t *testing.T) {
	type args struct {
		fileWAbsPath string
		content      string
	}
	type test struct {
		name    string
		args    args
		wantErr bool
	}
	tests := []test{
		{
			name: "success",
			args: args{
				fileWAbsPath: fmt.Sprintf("/tmp/%s", NewRandomString(10)),
				content:      "%s",
			},
			wantErr: false,
		},
		{
			name: "destination folder not exist",
			args: args{
				fileWAbsPath: fmt.Sprintf("/%s/%s", NewRandomString(10), NewRandomString(10)),
				content:      "%s",
			},
			wantErr: true,
		},
		{
			name: "empty content",
			args: args{
				fileWAbsPath: fmt.Sprintf("/%s/%s", NewRandomString(10), NewRandomString(10)),
				content:      "",
			},
			wantErr: true,
		},
	}

	// In CI env, root can do everything
	if os.Getenv("CI") == "" {
		tests = append(tests, test{
			name: "permission error",
			args: args{
				fileWAbsPath: fmt.Sprintf("/root/%s", NewRandomString(10)),
				content:      "%s",
			},
			wantErr: true,
		})
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewRegularFile(tt.args.fileWAbsPath, tt.args.content)

			if tt.wantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)

				_content, err := ioutil.ReadFile(tt.args.fileWAbsPath)
				assert.Nil(t, err)
				assert.Equal(t, string(_content), tt.args.content)

				err = os.Remove(tt.args.fileWAbsPath)
				assert.Nil(t, err)
			}

		})
	}
}

func TestFileDescriptor(t *testing.T) {
	fd := &FileDescription{
		Path:        "test",
		ContentType: "application/json",
		AllowedUser: "test",
		ExpiredAt:   time.Now().Add(1).Unix(),
	}

	sign, err := fd.Sign("test")
	if err != nil {
		t.Error(err)
		return
	}

	if !fd.Verify("test", sign) {
		t.Error("verify failed")
	}

	enc, err := fd.Encode("test")
	if err != nil {
		t.Error(err)
		return
	}

	dfd, err := DecodeDescriptor("test", enc)
	if err != nil {
		t.Error(err)
		return
	}

	if dfd.Path != fd.Path {
		t.Error("path not match")
	}

	if dfd.IsExpired() {
		t.Error("should not expired")
	}

	time.Sleep(1 * time.Second)
	if !dfd.IsExpired() {
		t.Error("should expired")
	}

	if !dfd.IsAllowedUser("test") {
		t.Error("should allowed")
	}

	if dfd.IsAllowedUser("test2") {
		t.Error("should not allowed")
	}
}
