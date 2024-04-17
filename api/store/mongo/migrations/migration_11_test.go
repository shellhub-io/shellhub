package migrations

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestMigration11(t *testing.T) {
	t.Cleanup(func() {
		assert.NoError(t, srv.Reset())
	})

	migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[:11]...)
	err := migrates.Up(context.Background(), migrate.AllAvailable)
	assert.NoError(t, err)

	pk := models.PrivateKey{
		CreatedAt: clock.Now(),
	}

	_, err = c.Database("test").Collection("private_keys").InsertOne(context.TODO(), pk)
	assert.NoError(t, err)

	index := c.Database("test").Collection("private_keys").Indexes()

	cursor, err := index.List(context.TODO())
	assert.NoError(t, err)

	var results []bson.M
	err = cursor.All(context.TODO(), &results)
	assert.NoError(t, err)

	keyField, ok := results[1]["key"].(primitive.M)
	if !ok {
		panic("type assertion failed")
	}

	assert.Equal(t, int32(1), keyField["created_at"])

	value, key := results[1]["expireAfterSeconds"]
	assert.Equal(t, true, key)
	assert.Equal(t, int32(60), value)
}
