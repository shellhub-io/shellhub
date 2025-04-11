package migrations

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	migrate "github.com/xakep666/mongo-migrate"
)

func TestMigration7(t *testing.T) {
	t.Cleanup(func() {
		assert.NoError(t, srv.Reset())
	})

	migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[:6]...)
	err := migrates.Up(context.Background(), migrate.AllAvailable)
	assert.NoError(t, err)

	recordedSession1 := models.RecordedSession{
		UID:     "uid",
		Message: "message",
	}

	recordedSession2 := models.RecordedSession{
		UID:     "uid",
		Message: "message",
	}

	_, err = c.Database("test").Collection("recorded_sessions").InsertOne(context.TODO(), recordedSession1)
	assert.NoError(t, err)

	_, err = c.Database("test").Collection("recorded_sessions").InsertOne(context.TODO(), recordedSession2)
	assert.NoError(t, err)

	migrates = migrate.NewMigrate(c.Database("test"), GenerateMigrations()[:7]...)
	err = migrates.Up(context.Background(), migrate.AllAvailable)
	assert.NoError(t, err)
}
