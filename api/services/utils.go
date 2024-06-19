package services

import (
	"os"

	jwt "github.com/golang-jwt/jwt/v4"
)

func LoadKeys(privateKey string) (*Keys, error) {
	signBytes, err := os.ReadFile(privateKey)
	if err != nil {
		return nil, err
	}

	privKey, err := jwt.ParseRSAPrivateKeyFromPEM(signBytes)
	if err != nil {
		return nil, err
	}

	return &Keys{
		PrivateKey: privKey,
		PublicKey:  &privKey.PublicKey,
	}, nil
}

func contains(list []string, item string) bool {
	for _, i := range list {
		if i == item {
			return true
		}
	}

	return false
}
