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

func TestMigration7(t *testing.T) {
	logrus.Info("Testing Migration 7 - Test if the uid and message is not unique")

	db := dbtest.DBServer{}
	defer db.Stop()

	migrates := migrate.NewMigrate(db.Client().Database("test"), GenerateMigrations()[:6]...)
	err := migrates.Up(migrate.AllAvailable)
	assert.NoError(t, err)

	recordedSession1 := models.RecordedSession{
		UID:     "uid",
		Message: "message",
	}

	recordedSession2 := models.RecordedSession{
		UID:     "uid",
		Message: "message",
	}

	_, err = db.Client().Database("test").Collection("recorded_sessions").InsertOne(context.TODO(), recordedSession1)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("recorded_sessions").InsertOne(context.TODO(), recordedSession2)
	assert.NoError(t, err)

	migrates = migrate.NewMigrate(db.Client().Database("test"), GenerateMigrations()[:7]...)
	err = migrates.Up(migrate.AllAvailable)
	assert.NoError(t, err)
}
