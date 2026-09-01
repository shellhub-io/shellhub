package services

import (
	"crypto/rsa"
	"os"

	jwt "github.com/golang-jwt/jwt/v5"
)

// LoadKeys reads the RSA key pair the API signs its tokens with, from the paths named by
// PRIVATE_KEY and PUBLIC_KEY.
func LoadKeys() (*rsa.PrivateKey, *rsa.PublicKey, error) {
	signBytes, err := os.ReadFile(os.Getenv("PRIVATE_KEY")) //nolint:gosec // G703: path comes from trusted env var
	if err != nil {
		return nil, nil, err
	}

	privKey, err := jwt.ParseRSAPrivateKeyFromPEM(signBytes)
	if err != nil {
		return nil, nil, err
	}

	verifyBytes, err := os.ReadFile(os.Getenv("PUBLIC_KEY")) //nolint:gosec // G703: path comes from trusted env var
	if err != nil {
		return nil, nil, err
	}

	pubKey, err := jwt.ParseRSAPublicKeyFromPEM(verifyBytes)
	if err != nil {
		return nil, nil, err
	}

	return privKey, pubKey, nil
}
