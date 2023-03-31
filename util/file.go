package util

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"strings"
	"time"

	b58 "github.com/jbenet/go-base58"
)

const (
	ErrInvalidSignature = StorageError("invalid signature")
	ErrExpiredUrl       = StorageError("url is expired")
)

type StorageError string

func (e StorageError) Error() string { return string(e) }

func (e StorageError) ToJSON() string {
	return `{"error": "` + e.Error() + `"}`
}

// NewRegularFile creates regular file with specified content
func NewRegularFile(fileWAbsPath string, content string) error {
	permission := 0755
	err := os.WriteFile(fileWAbsPath, //nolint:gosec
		[]byte(content), fs.FileMode(permission))
	if err != nil {
		return err
	}

	return nil
}

type FileDescription struct {
	Path        string `json:"path"`
	ContentType string `json:"content_type"`
	AllowedUser string `json:"allowed_user"`
	ExpiredAt   int64  `json:"expired_at"`
}

func (f *FileDescription) Sign(key string) (string, error) {

	b, err := json.Marshal(f)
	if err != nil {
		return "", err
	}

	mac := hmac.New(sha256.New, []byte(key))
	mac.Write(b)

	return b58.Encode(mac.Sum(nil)), nil
}

func (f *FileDescription) Verify(key, sign string) bool {
	s, err := f.Sign(key)
	if err != nil {
		return false
	}
	return s == sign
}

func (f *FileDescription) IsExpired() bool {
	return f.ExpiredAt != 0 && f.ExpiredAt < time.Now().Unix()
}

func (f *FileDescription) IsAllowedUser(user string) bool {
	return f.AllowedUser == "" || strings.Contains(f.AllowedUser, user)
}

func (f *FileDescription) Encode(key string) (string, error) {
	b, err := json.Marshal(f)
	if err != nil {
		return "", err
	}

	if key == "" {
		return base64.RawStdEncoding.EncodeToString(b), nil
	}

	sign, err := f.Sign(key)
	if err != nil {
		return "", err
	}

	return base64.RawStdEncoding.EncodeToString(b) + "." + sign, nil
}

func DecodeDescriptor(key, desc string) (*FileDescription, error) {
	parts := strings.Split(desc, ".")
	if len(parts) == 1 && key != "" {
		return nil, errors.New("missing signature")
	}

	b, err := base64.RawStdEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, err
	}

	var f FileDescription
	if err := json.Unmarshal(b, &f); err != nil {
		return nil, err
	}

	if key == "" {
		return &f, nil
	}

	if !f.Verify(key, parts[1]) {
		return nil, ErrInvalidSignature
	}

	return &f, nil
}
