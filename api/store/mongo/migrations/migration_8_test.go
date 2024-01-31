package migrations

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	migrate "github.com/xakep666/mongo-migrate"
)

func TestMigration8(t *testing.T) {
	logrus.Info("Testing Migration 8 - Test if the recorded is not unique")

	db := dbtest.DB{}
	err := func() error {
		err := db.Down(context.Background())

		return err
	}()
	assert.NoError(t, err)

	migrates := migrate.NewMigrate(db.MongoClient.Database("test"), GenerateMigrations()[:7]...)
	err = migrates.Up(context.Background(), migrate.AllAvailable)
	assert.NoError(t, err)

	session1 := models.Session{
		Recorded: true,
	}

	session2 := models.Session{
		Recorded: true,
	}

	_, err = mongoClient.Database("test").Collection("sessions").InsertOne(context.TODO(), session1)
	assert.NoError(t, err)

	_, err = mongoClient.Database("test").Collection("sessions").InsertOne(context.TODO(), session2)
	assert.NoError(t, err)

	migrates = migrate.NewMigrate(mongoClient.Database("test"), GenerateMigrations()[:8]...)
	err = migrates.Up(context.Background(), migrate.AllAvailable)
	assert.NoError(t, err)
}
