package migrations

import (
	"context"
	"os"
	"testing"

	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	mongooptions "go.mongodb.org/mongo-driver/mongo/options"
)

var (
	srv = &dbtest.Server{}
	c   *mongo.Client
)

func TestMain(m *testing.M) {
	os.Setenv("SHELLHUB_ENTERPRISE", "true")
	os.Setenv("SHELLHUB_CLOUD", "true")

	log.Info("Starting migration tests")

	ctx := context.Background()

	srv.Container.Database = "test"

	if err := srv.Up(ctx); err != nil {
		log.WithError(err).Error("Failed to UP the mongodb container")
		os.Exit(1)
	}

	log.Info("Connecting to ", srv.Container.ConnectionString)

	var err error

	c, err = mongo.Connect(ctx, mongooptions.Client().ApplyURI(srv.Container.ConnectionString+"/"+srv.Container.Database))
	if err != nil {
		log.WithError(err).Error("Unable to connect to MongoDB")
		os.Exit(1)
	}

	if err := c.Ping(ctx, nil); err != nil {
		log.WithError(err).Error("Unable to ping MongoDB")
		os.Exit(1)
	}

	code := m.Run()

	log.Info("Stopping migration tests")
	if err := srv.Down(ctx); err != nil {
		log.WithError(err).Error("Failed to DOWN the mongodb container")
		os.Exit(1)
	}

	os.Exit(code)
}
