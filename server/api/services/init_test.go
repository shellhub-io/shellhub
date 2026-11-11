package services

import (
	"crypto/rand"
	"crypto/rsa"
	"os"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/server/api/store/pg/dbtest"
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
	now = clock.New().Now()
	os.Exit(dbtest.Run(m))
}
