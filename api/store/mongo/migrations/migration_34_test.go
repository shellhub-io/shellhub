package migrations

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestMigration34(t *testing.T) {
	t.Cleanup(func() {
		assert.NoError(t, srv.Reset())
	})

	migrations := GenerateMigrations()[:33]

	migrates := migrate.NewMigrate(c.Database("test"), migrations...)

	err := migrates.Up(context.Background(), migrate.AllAvailable)
	assert.NoError(t, err)

	version, _, err := migrates.Version(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, uint64(33), version)

	migrations = GenerateMigrations()[:34]

	migrates = migrate.NewMigrate(c.Database("test"), migrations...)
	err = migrates.Up(context.Background(), migrate.AllAvailable)
	assert.NoError(t, err)

	version, _, err = migrates.Version(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, uint64(34), version)

	cursor, err := c.Database("test").Collection("devices").Indexes().List(context.TODO())
	assert.NoError(t, err)

	var results []bson.M
	err = cursor.All(context.TODO(), &results)
	assert.NoError(t, err)

	indexes := []string{}

	for _, index := range results {
		if v, ok := index["key"].(primitive.M); ok {
			for key := range v {
				indexes = append(indexes, key)
			}
		}
	}

	assert.Contains(t, indexes, "online")
}
