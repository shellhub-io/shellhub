package services

import (
	"crypto/rand"
	"crypto/rsa"
	"os"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/clock"
	clockmocks "github.com/shellhub-io/shellhub/pkg/clock/mocks"
	"github.com/shellhub-io/shellhub/pkg/envs"
	env_mocks "github.com/shellhub-io/shellhub/pkg/envs/mocks"
	"github.com/shellhub-io/shellhub/pkg/hash"
	hashmock "github.com/shellhub-io/shellhub/pkg/hash/mocks"
	"github.com/shellhub-io/shellhub/pkg/uuid"
)

var (
	privateKey      *rsa.PrivateKey
	publicKey       *rsa.PublicKey
	envMock         *env_mocks.MockBackend
	clockMock       *clockmocks.MockClock
	hashMock        *hashmock.MockHasher
	now             time.Time
	realUUIDBackend uuid.UUID
)

func TestMain(m *testing.M) {
	privateKey, _ = rsa.GenerateKey(rand.Reader, 2048)
	publicKey = &privateKey.PublicKey
	now = clock.Now()
	realUUIDBackend = uuid.DefaultBackend
	clockMock = &clockmocks.MockClock{}
	envMock = &env_mocks.MockBackend{}
	clock.DefaultBackend = clockMock
	envs.DefaultBackend = envMock
	hashMock = &hashmock.MockHasher{}
	hash.Backend = hashMock
	code := m.Run()
	os.Exit(code)
}
