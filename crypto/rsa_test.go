package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRSA(t *testing.T) {
	msg := "message"
	sk, pk, err := GenerateKeyPair(2048)
	assert.Nil(t, err)
	assert.NotNil(t, sk)
	assert.NotNil(t, pk)

	pb, err := PublicKeyToBytes(pk)
	assert.Nil(t, err)

	rpk, err := BytesToPublicKey(pb)
	assert.Nil(t, err)
	assert.Equal(t, pk, rpk)

	c, err := EncryptWithPublicKey([]byte(msg), pk)
	assert.Nil(t, err)
	assert.NotNil(t, c)

	d, err := DecryptWithPrivateKey(c, sk)
	assert.Nil(t, err)
	assert.Equal(t, msg, string(d))
}
