package util

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"os"

	b58 "github.com/jbenet/go-base58"
)

//Hash calculate value hash
func Hash(v interface{}) []byte {
	var b []byte
	switch v := v.(type) {
	case []byte:
		b = v
	default:
		b = []byte(fmt.Sprintf("%v", v))
	}

	h := sha256.Sum256(b)
	return h[:]
}

func Hash64(v interface{}) string {
	h := Hash(v)
	return base64.StdEncoding.EncodeToString(h[:])
}

func Hash58(v interface{}) string {
	h := Hash(v)
	return b58.Encode(h[:])
}

func HashHex(v interface{}) string {
	h := Hash(v)
	return hex.EncodeToString(h[:])
}

func HashFile(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}

func HashFile64(path string) (string, error) {
	h, err := HashFile(path)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(h), nil
}
