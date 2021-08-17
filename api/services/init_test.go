package services

import (
	"crypto/rand"
	"crypto/rsa"
	"os"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/clock"
	clockmocks "github.com/shellhub-io/shellhub/pkg/clock/mocks"
	"github.com/shellhub-io/shellhub/pkg/requests/mocks"
)

var (
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	clientMock *mocks.Client
	clockMock  *clockmocks.Clock
	now        time.Time
)

func TestMain(m *testing.M) {
	privateKey, _ = rsa.GenerateKey(rand.Reader, 2048)
	publicKey = &privateKey.PublicKey
	clientMock = &mocks.Client{}
	clockMock = &clockmocks.Clock{}
	clock.DefaultBackend = clockMock
	now = time.Now()
	code := m.Run()
	os.Exit(code)
}
