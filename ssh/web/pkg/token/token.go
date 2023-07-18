// Package token provides a interface to create and parse session's token.
package token

import (
	"crypto/rsa"
	"fmt"

	"github.com/golang-jwt/jwt"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	"github.com/shellhub-io/shellhub/ssh/pkg/magickey"
)

// Token represents a web session's token.
type Token struct {
	// ID is a UUID used to identify the token.
	// It is used to retrieve the data from the cache.
	ID string
	// Data is a JWT token.
	Data string
}

// NewToken creates a new token.
func NewToken(_ *rsa.PrivateKey) (*Token, error) {
	identifier := uuid.Generate()

	token, err := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"id": identifier,
	}).SignedString(magickey.GetRerefence())
	if err != nil {
		return nil, err
	}

	return &Token{ID: identifier, Data: token}, nil
}

// Parse a JWT token to a session's token.
func Parse(token string) (*Token, error) {
	claims := new(jwt.MapClaims)
	if _, err := jwt.ParseWithClaims(token, claims, func(jwtToken *jwt.Token) (interface{}, error) {
		if _, ok := jwtToken.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected method: %s", jwtToken.Header["alg"])
		}

		return magickey.GetRerefence().Public().(*rsa.PublicKey), nil
	}); err != nil {
		return nil, err
	}

	id := (*claims)["id"].(string) //nolint: forcetypeassert

	return &Token{ID: id, Data: token}, nil
}
