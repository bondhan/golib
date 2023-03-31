package authtoken

import "github.com/golang-jwt/jwt/v4"

type ClaimsWithRoles struct {
	jwt.RegisteredClaims
	Roles []string `json:"roles"`
}
