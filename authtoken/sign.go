package authtoken

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// Issuer    string    `json:"iss,omitempty"`
// Subject   string    `json:"sub,omitempty"`
// Audience  []string  `json:"aud,omitempty"`
// ExpiresAt time.Time `json:"exp,omitempty"`
// NotBefore time.Time `json:"nbf,omitempty"`
// IssuedAt  time.Time `json:"iat,omitempty"`

func sign(duration time.Duration, issuer, subject string, roles, aud []string, key interface{}) (string, error) {

	exp := time.Now().Add(duration)
	now := time.Now()

	claimsWithRoles := ClaimsWithRoles{
		Roles: roles,
		RegisteredClaims: jwt.RegisteredClaims{
			Audience:  aud,
			ExpiresAt: jwt.NewNumericDate(exp),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    issuer,
			NotBefore: jwt.NewNumericDate(now),
			Subject:   subject,
		},
	}

	jwtWithClaims := jwt.NewWithClaims(jwt.SigningMethodRS512, claimsWithRoles)

	token, err := jwtWithClaims.SignedString(key)
	if err != nil {
		return "", err
	}

	return token, nil
}
