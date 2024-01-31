package migrations

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
)

func TestMigration52(t *testing.T) {
	logrus.Info("Testing Migration 52")

	db := dbtest.DB{}
	err := func() error {
		err := db.Down(context.Background())

		return err
	}()
	assert.NoError(t, err)

	user := models.User{}

	_, err = mongoClient.Database("test").Collection("users").InsertOne(context.Background(), user)
	assert.NoError(t, err)

	cases := []struct {
		description string
		Test        func(t *testing.T)
	}{
		{
			"Success to apply up on migration 52",
			func(t *testing.T) {
				t.Helper()

				migrations := GenerateMigrations()[51:52]
				migrates := migrate.NewMigrate(mongoClient.Database("test"), migrations...)
				err = migrates.Up(context.Background(), migrate.AllAvailable)
				assert.NoError(t, err)

				key := new(models.User)
				result := mongoClient.Database("test").Collection("users").FindOne(context.Background(), bson.M{})
				assert.NoError(t, result.Err())

				err = result.Decode(key)
				assert.NoError(t, err)

				assert.True(t, key.EmailMarketing)
			},
		},
		{
			"Success to apply down on migration 52",
			func(t *testing.T) {
				t.Helper()

				migrations := GenerateMigrations()[51:52]
				migrates := migrate.NewMigrate(mongoClient.Database("test"), migrations...)
				err := migrates.Down(context.Background(), migrate.AllAvailable)
				assert.NoError(t, err)

				key := new(models.User)
				result := mongoClient.Database("test").Collection("users").FindOne(context.Background(), bson.M{})
				assert.NoError(t, result.Err())

				err = result.Decode(key)
				assert.NoError(t, err)

				assert.False(t, key.EmailMarketing)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, tc.Test)
	}
}
