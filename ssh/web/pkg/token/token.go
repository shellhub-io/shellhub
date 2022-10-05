package token

import (
	"crypto/rsa"
	"fmt"

	"github.com/golang-jwt/jwt"
	"github.com/shellhub-io/shellhub/ssh/pkg/magickey"
)

type Token struct {
	ID    string
	Token string
}

func NewToken(id string, key *rsa.PrivateKey) (*Token, error) {
	token, err := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"id": id,
	}).SignedString(magickey.GetRerefence())
	if err != nil {
		return nil, err
	}

	return &Token{ID: id, Token: token}, nil
}

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

	id := (*claims)["id"].(string) // nolint: forcetypeassert

	return &Token{ID: id, Token: token}, nil
}
