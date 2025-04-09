package migrations

import (
	"context"
	"os"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
)

func TestMigration47(t *testing.T) {
	cases := []struct {
		description string
		Test        func(t *testing.T)
	}{
		{
			"Success to apply up on migration 47",
			func(t *testing.T) {
				t.Helper()

				sessionWithoutPossition := &models.Session{
					UID:       "test",
					IPAddress: "201.182.197.68",
				}

				_, err := c.Database("test").Collection("sessions").InsertOne(context.Background(), sessionWithoutPossition)
				assert.NoError(t, err)

				migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[46:47]...)
				assert.NoError(t, migrates.Up(context.Background(), migrate.AllAvailable))

				key := new(models.Session)
				result := c.Database("test").Collection("sessions").FindOne(context.Background(), bson.M{"uid": sessionWithoutPossition.UID})
				assert.NoError(t, result.Err())

				err = result.Decode(key)
				assert.NoError(t, err)

				if os.Getenv("GEOIP") == "true" {
					assert.NotEqual(t, sessionWithoutPossition.Position, key.Position)
				} else {
					assert.Equal(t, sessionWithoutPossition.Position, key.Position)
				}
			},
		},
		{
			"Success to apply down on migration 47",
			func(t *testing.T) {
				t.Helper()

				sessionWithoutPossition := &models.Session{
					UID:       "test",
					IPAddress: "201.182.197.68",
				}

				_, err := c.Database("test").Collection("sessions").InsertOne(context.Background(), sessionWithoutPossition)
				assert.NoError(t, err)

				migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[46:47]...)
				assert.NoError(t, migrates.Down(context.Background(), migrate.AllAvailable))

				key := new(models.Session)
				result := c.Database("test").Collection("sessions").FindOne(context.Background(), bson.M{"uid": sessionWithoutPossition.UID})
				assert.NoError(t, result.Err())

				err = result.Decode(key)
				assert.NoError(t, err)

				assert.Equal(t, sessionWithoutPossition.Position, key.Position)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			t.Cleanup(func() {
				assert.NoError(t, srv.Reset())
			})
			tc.Test(t)
		})
	}
}
