package jwttoken

import (
	"crypto/rsa"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

func encodeClaims(claims jwt.Claims, privateKey *rsa.PrivateKey) (string, error) {
	return jwt.NewWithClaims(jwt.SigningMethodRS256, claims).SignedString(privateKey)
}

func decodeClaims[T jwt.Claims](publicKey *rsa.PublicKey, raw string, claims T) error {
	_, err := jwt.ParseWithClaims(raw, claims, evalClaims(publicKey), jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Alg()}))

	return err
}

func evalClaims(publicKey *rsa.PublicKey) jwt.Keyfunc {
	return func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signature method: %v", t.Header["alg"])
		}

		return publicKey, nil
	}
}
