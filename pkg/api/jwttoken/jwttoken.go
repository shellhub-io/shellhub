package jwttoken

import (
	"crypto/rsa"

	"github.com/golang-jwt/jwt/v4"
)

// Encode encodes the provided claims into a JWT token using the provided RSA private key.
// It returns the encoded JWT token as a string and any error encountered during the encoding process.
// The claims are signed using the RS256 signing method.
func Encode(claims jwt.Claims, privateKey *rsa.PrivateKey) (string, error) {
	return jwt.NewWithClaims(jwt.SigningMethodRS256, claims).SignedString(privateKey)
}
