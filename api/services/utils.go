package services

import (
	"crypto/rsa"
	"io/ioutil"
	"os"

	"github.com/golang-jwt/jwt/v4"
)

func LoadKeys() (*rsa.PrivateKey, *rsa.PublicKey, error) {
	signBytes, err := ioutil.ReadFile(os.Getenv("PRIVATE_KEY"))
	if err != nil {
		return nil, nil, err
	}

	privKey, err := jwt.ParseRSAPrivateKeyFromPEM(signBytes)
	if err != nil {
		return nil, nil, err
	}

	verifyBytes, err := ioutil.ReadFile(os.Getenv("PUBLIC_KEY"))
	if err != nil {
		return nil, nil, err
	}

	pubKey, err := jwt.ParseRSAPublicKeyFromPEM(verifyBytes)
	if err != nil {
		return nil, nil, err
	}

	return privKey, pubKey, nil
}

func contains(list []string, item string) bool {
	for _, i := range list {
		if i == item {
			return true
		}
	}

	return false
}
