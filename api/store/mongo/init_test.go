package mongo

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	"github.com/shellhub-io/shellhub/pkg/cache"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	mongoClient *mongo.Client
	mongostore  *Store
	mongoHost   string
)

func TestMain(m *testing.M) {
	var code = 1
	defer func() { os.Exit(code) }()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()
	container, err := dbtest.Setup(ctx)
	if err != nil {
		log.Fatalf("failed to setup test container: %s", err)
	}
	defer container.Stop()

	mongostore = NewStore(container.Client.Database("test"), cache.NewNullCache())
	mongoClient = container.Client
	mongoHost = container.Host

	err = mongoClient.Database("test").Drop(ctx)
	if err != nil {
		log.Fatalf("Failed to drop database: %s", err)
	}

	code = m.Run()
}

func GetMongoStore() *Store {
	return mongostore
}
