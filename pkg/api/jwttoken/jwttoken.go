package jwttoken

import (
	"crypto/rsa"
	"fmt"

	"github.com/golang-jwt/jwt/v4"
)

// Encode encodes the provided claims into a JWT token using the provided RSA private key.
// It returns the encoded JWT token as a string and any error encountered during the encoding process.
// The claims are signed using the RS256 signing method.
func Encode(claims jwt.Claims, privateKey *rsa.PrivateKey) (string, error) {
	return jwt.NewWithClaims(jwt.SigningMethodRS256, claims).SignedString(privateKey)
}

// Decode decodes the raw JWT to claims.
func Decode[T jwt.Claims](publicKey *rsa.PublicKey, raw string, claims T) error {
	_, err := jwt.ParseWithClaims(raw, claims, eval(publicKey), jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Alg()}))

	return err
}

// eval evaluates if a token t is a valid token.
func eval(publicKey *rsa.PublicKey) jwt.Keyfunc {
	return func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signature method: %v", t.Header["alg"])
		}

		return publicKey, nil
	}
}
