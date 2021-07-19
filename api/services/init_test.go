package services

import (
	"crypto/rand"
	"crypto/rsa"
	"os"
	"testing"
)

var (
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
)

func TestMain(m *testing.M) {
	privateKey, _ = rsa.GenerateKey(rand.Reader, 2048)
	publicKey = &privateKey.PublicKey
	code := m.Run()
	os.Exit(code)
}
