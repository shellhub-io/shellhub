package migrations

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	mongoClient *mongo.Client
)

func TestMain(m *testing.M) {
	os.Setenv("SHELLHUB_ENTERPRISE", "true")
	os.Setenv("SHELLHUB_CLOUD", "true")

	ctx := context.Background()
	container, err := dbtest.Setup(ctx)
	if err != nil {
		log.Fatalf("failed to setup test container: %s", err)
	}
	defer container.Stop()

	mongoClient = container.Client

	code := m.Run()
	os.Exit(code)
}
