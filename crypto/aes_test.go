package crypto

import (
	"crypto/sha256"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAES(t *testing.T) {
	ks := "pwd"
	msg := "message"
	h := sha256.Sum256([]byte(ks))
	c, err := AESEncrypt([]byte(msg), h[:])
	assert.Nil(t, err)
	assert.NotNil(t, c)

	d, err := AESDecrypt(c, h[:])
	assert.Nil(t, err)
	assert.Equal(t, msg, string(d))
}
