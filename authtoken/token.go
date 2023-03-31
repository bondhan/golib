package authtoken

import (
	"crypto/rsa"
	"time"
)

// SignToken signs the token with the private key
// duration is the duration of the token in seconds
func SignToken(roles []string, aud, userID string, duration time.Duration, key *rsa.PrivateKey) (*Session, error) {
	var audience []string

	audience = append(audience, aud)

	access, err := sign(duration, "auth-service", userID, roles, audience, key)
	if err != nil {
		return nil, err
	}

	claims, err := verify(access, &key.PublicKey)
	if err != nil {
		return nil, err
	}

	return &Session{
		Token:   access,
		Expires: claims.ExpiresAt.Unix(),
		User:    userID,
	}, nil
}

// GetClaims get Claims data from string token and key
func GetClaims(token string, key *rsa.PrivateKey) (*ClaimsWithRoles, error) {
	claims, err := verify(token, &key.PublicKey)
	if err != nil {
		return nil, err
	}

	return claims, err
}

type Session struct {
	Expires int64
	Token   string
	User    string
}
