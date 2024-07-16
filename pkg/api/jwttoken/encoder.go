package jwttoken

import (
	"crypto/rsa"
	"fmt"

	"github.com/golang-jwt/jwt/v4"
)

// encodeClaims encodes the provided claims into a JWT token using the provided RSA private key.
// It returns the encoded JWT token as a string and any error encountered during the encoding process.
// The claims are signed using the RS256 signing method.
func encodeClaims(claims jwt.Claims, privateKey *rsa.PrivateKey) (string, error) {
	return jwt.NewWithClaims(jwt.SigningMethodRS256, claims).SignedString(privateKey)
}

// decodeClaims decodes the raw JWT into claims.
func decodeClaims[T jwt.Claims](publicKey *rsa.PublicKey, raw string, claims T) error {
	_, err := jwt.ParseWithClaims(raw, claims, evalClaims(publicKey), jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Alg()}))

	return err
}

// evalClaims evaluates if a token is valid.
func evalClaims(publicKey *rsa.PublicKey) jwt.Keyfunc {
	return func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signature method: %v", t.Header["alg"])
		}

		return publicKey, nil
	}
}
