package authtoken

import (
	"crypto/rsa"
	"fmt"
	"strings"

	"github.com/golang-jwt/jwt/v4"
)

func ParseHeader(raw string) (string, error) {
	has := strings.HasPrefix(raw, "Bearer ")
	if !has {
		return "", fmt.Errorf("invalid JWT Token")
	}
	split := strings.Split(raw, "Bearer ")

	return split[1], nil
}

func verify(tokenStr string, key *rsa.PublicKey) (*ClaimsWithRoles, error) {

	token, err := jwt.ParseWithClaims(tokenStr, &ClaimsWithRoles{}, func(t *jwt.Token) (interface{}, error) {
		return key, nil
	})

	if err != nil {
		return nil, err
	}

	claims := token.Claims.(*ClaimsWithRoles)

	err = claims.Valid()
	if err != nil {
		return nil, err
	}

	return claims, nil
}
