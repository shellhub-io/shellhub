package services

import (
	"crypto/rsa"
	"os"

	jwt "github.com/golang-jwt/jwt/v5"
)

// LoadKeys reads the RSA key pair the API signs its tokens with from the PEM files at privatePath
// and publicPath. It returns the error of the first file it cannot read or parse.
func LoadKeys(privatePath, publicPath string) (*rsa.PrivateKey, *rsa.PublicKey, error) {
	signBytes, err := os.ReadFile(privatePath) //nolint:gosec // G304: the path comes from the operator's configuration
	if err != nil {
		return nil, nil, err
	}

	privKey, err := jwt.ParseRSAPrivateKeyFromPEM(signBytes)
	if err != nil {
		return nil, nil, err
	}

	verifyBytes, err := os.ReadFile(publicPath) //nolint:gosec // G304: the path comes from the operator's configuration
	if err != nil {
		return nil, nil, err
	}

	pubKey, err := jwt.ParseRSAPublicKeyFromPEM(verifyBytes)
	if err != nil {
		return nil, nil, err
	}

	return privKey, pubKey, nil
}
