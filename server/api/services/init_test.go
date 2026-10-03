package services

import (
	"crypto/rand"
	"crypto/rsa"
	"os"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/clock"
)

const testIssuer = "http://localhost"

var (
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	now        time.Time
)

func TestMain(m *testing.M) {
	privateKey, _ = rsa.GenerateKey(rand.Reader, 2048)
	publicKey = &privateKey.PublicKey
	now = clock.Now()
	code := m.Run()
	os.Exit(code)
}
