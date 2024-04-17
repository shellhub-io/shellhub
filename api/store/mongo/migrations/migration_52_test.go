package migrations

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/api/pkg/fixtures"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
)

func TestMigration52(t *testing.T) {
	user := models.User{}

	cases := []struct {
		description string
		Test        func(t *testing.T)
	}{
		{
			"Success to apply up on migration 52",
			func(t *testing.T) {
				t.Helper()

				_, err := srv.Client().Database("test").Collection("users").InsertOne(context.Background(), user)
				assert.NoError(t, err)

				migrates := migrate.NewMigrate(srv.Client().Database("test"), GenerateMigrations()[51:52]...)
				assert.NoError(t, migrates.Up(context.Background(), migrate.AllAvailable))

				key := new(models.User)
				result := srv.Client().Database("test").Collection("users").FindOne(context.Background(), bson.M{})
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

				_, err := srv.Client().Database("test").Collection("users").InsertOne(context.Background(), user)
				assert.NoError(t, err)

				migrates := migrate.NewMigrate(srv.Client().Database("test"), GenerateMigrations()[51:52]...)
				assert.NoError(t, migrates.Down(context.Background(), migrate.AllAvailable))

				key := new(models.User)
				result := srv.Client().Database("test").Collection("users").FindOne(context.Background(), bson.M{})
				assert.NoError(t, result.Err())

				err = result.Decode(key)
				assert.NoError(t, err)

				assert.False(t, key.EmailMarketing)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			t.Cleanup(func() {
				assert.NoError(t, fixtures.Teardown())
			})
			tc.Test(t)
		})
	}
}
