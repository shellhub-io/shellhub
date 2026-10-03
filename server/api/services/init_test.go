package services

import (
	"crypto/rand"
	"crypto/rsa"
	"os"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/hash"
	hashmock "github.com/shellhub-io/shellhub/pkg/hash/mocks"
)

const testIssuer = "http://localhost"

var (
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	hashMock   *hashmock.MockHasher
	now        time.Time
)

func TestMain(m *testing.M) {
	privateKey, _ = rsa.GenerateKey(rand.Reader, 2048)
	publicKey = &privateKey.PublicKey
	now = clock.Now()
	hashMock = &hashmock.MockHasher{}
	hash.Backend = hashMock
	code := m.Run()
	os.Exit(code)
}
