package migrations

import (
	"context"
	"os"
	"testing"

	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
)

func TestMigration47(t *testing.T) {
	logrus.Info("Testing Migration 47")

	db := dbtest.DBServer{}
	defer db.Stop()

	sessionWithoutPossition := &models.Session{
		UID:       "test",
		IPAddress: "201.182.197.68",
	}

	_, err := db.Client().Database("test").Collection("sessions").InsertOne(context.Background(), sessionWithoutPossition)
	assert.NoError(t, err)

	cases := []struct {
		description string
		Test        func(t *testing.T)
	}{
		{
			"Success to apply up on migration 47",
			func(t *testing.T) {
				t.Helper()

				migrations := GenerateMigrations()[46:47]
				migrates := migrate.NewMigrate(db.Client().Database("test"), migrations...)
				err := migrates.Up(context.Background(), migrate.AllAvailable)
				assert.NoError(t, err)

				key := new(models.Session)
				result := db.Client().Database("test").Collection("sessions").FindOne(context.Background(), bson.M{"uid": sessionWithoutPossition.UID})
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

				migrations := GenerateMigrations()[46:47]
				migrates := migrate.NewMigrate(db.Client().Database("test"), migrations...)
				err := migrates.Down(context.Background(), migrate.AllAvailable)
				assert.NoError(t, err)

				key := new(models.Session)
				result := db.Client().Database("test").Collection("sessions").FindOne(context.Background(), bson.M{"uid": sessionWithoutPossition.UID})
				assert.NoError(t, result.Err())

				err = result.Decode(key)
				assert.NoError(t, err)

				assert.Equal(t, sessionWithoutPossition.Position, key.Position)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, tc.Test)
	}
}
