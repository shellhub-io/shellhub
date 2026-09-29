// Package token provides a interface to create and parse session's token.
package token

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	"github.com/shellhub-io/shellhub/server/ssh/pkg/magickey"
)

// Token represents a web session's token.
type Token struct {
	// ID is a UUID used to identify the token.
	// It is used to retrieve the data from the cache.
	ID string
	// Data is a JWT token.
	Data string
}

// NewToken creates a new token, signed with the process-local key from [magickey.GetReference]
// rather than with any key the caller holds, so only this process can [Parse] it. issuer names
// the instance signing it and is written to the iss claim verbatim.
func NewToken(issuer string) (*Token, error) {
	identifier := uuid.Generate()

	token, err := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"id":  identifier,
		"iss": issuer,
	}).SignedString(magickey.GetReference())
	if err != nil {
		return nil, err
	}

	return &Token{ID: identifier, Data: token}, nil
}
