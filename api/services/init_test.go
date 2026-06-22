package services

import (
	"crypto/rand"
	"crypto/rsa"
	"os"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/internalclient/mocks"
	"github.com/shellhub-io/shellhub/pkg/clock"
	clockmocks "github.com/shellhub-io/shellhub/pkg/clock/mocks"
	"github.com/shellhub-io/shellhub/pkg/envs"
	env_mocks "github.com/shellhub-io/shellhub/pkg/envs/mocks"
	"github.com/shellhub-io/shellhub/pkg/hash"
	hashmock "github.com/shellhub-io/shellhub/pkg/hash/mocks"
)

var (
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	clientMock *mocks.MockClient
	envMock    *env_mocks.MockBackend
	clockMock  *clockmocks.MockClock
	hashMock   *hashmock.MockHasher
	now        time.Time
)

func TestMain(m *testing.M) {
	privateKey, _ = rsa.GenerateKey(rand.Reader, 2048)
	publicKey = &privateKey.PublicKey
	// Capture now before swapping clock.DefaultBackend so that clock.Now()
	// uses the real wall-clock backend and returns a valid timestamp.
	now = clock.Now()
	clientMock = &mocks.MockClient{}
	clockMock = &clockmocks.MockClock{}
	envMock = &env_mocks.MockBackend{}
	clock.DefaultBackend = clockMock
	envs.DefaultBackend = envMock
	hashMock = &hashmock.MockHasher{}
	hash.Backend = hashMock
	code := m.Run()
	os.Exit(code)
}
