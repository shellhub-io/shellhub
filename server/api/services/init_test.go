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
	clientMock *mocks.Client
	envMock    *env_mocks.Backend
	clockMock  *clockmocks.Clock
	hashMock   *hashmock.Hasher
	now        time.Time
)

func TestMain(m *testing.M) {
	privateKey, _ = rsa.GenerateKey(rand.Reader, 2048)
	publicKey = &privateKey.PublicKey
	clientMock = &mocks.Client{}
	clockMock = &clockmocks.Clock{}
	envMock = &env_mocks.Backend{}
	clock.DefaultBackend = clockMock
	envs.DefaultBackend = envMock
	hashMock = &hashmock.Hasher{}
	hash.Backend = hashMock
	now = time.Now()
	code := m.Run()
	os.Exit(code)
}
