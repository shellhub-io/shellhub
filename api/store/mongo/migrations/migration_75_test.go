package migrations

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/envs"
	envmock "github.com/shellhub-io/shellhub/pkg/envs/mocks"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestMigration75Up(t *testing.T) {
	ctx := context.Background()

	mock := &envmock.Backend{}
	envs.DefaultBackend = mock

	cases := []struct {
		description string
		setup       func(primitive.ObjectID) error
		expected    string
	}{
		{
			description: "Success to apply up on migration 75 with confirmed == true",
			setup: func(objID primitive.ObjectID) error {
				_, err := c.
					Database("test").
					Collection("users").
					InsertOne(ctx, map[string]interface{}{
						"_id":       objID,
						"confirmed": true,
					})

				return err
			},
			expected: models.UserStatusConfirmed.String(),
		},
		{
			description: "Success to apply up on migration 75 with confirmed == false",
			setup: func(objID primitive.ObjectID) error {
				_, err := c.
					Database("test").
					Collection("users").
					InsertOne(ctx, map[string]interface{}{
						"_id":       objID,
						"confirmed": false,
					})

				return err
			},
			expected: models.UserStatusNotConfirmed.String(),
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			tt.Cleanup(func() {
				assert.NoError(tt, srv.Reset())
			})

			objID := primitive.NewObjectID()
			assert.NoError(tt, tc.setup(objID))

			migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[74])
			require.NoError(tt, migrates.Up(context.Background(), migrate.AllAvailable))

			query := c.
				Database("test").
				Collection("users").
				FindOne(context.TODO(), bson.M{"_id": objID})

			user := make(map[string]interface{})
			require.NoError(tt, query.Decode(&user))

			_, ok := user["confirmed"]
			require.Equal(tt, false, ok)

			status, ok := user["status"].(string)
			require.Equal(tt, true, ok)
			require.Equal(tt, tc.expected, status)
		})
	}
}

func TestMigration75Down(t *testing.T) {
	ctx := context.Background()

	mock := &envmock.Backend{}
	envs.DefaultBackend = mock

	cases := []struct {
		description string
		setup       func(primitive.ObjectID) error
		expected    bool
	}{
		{
			description: "Success to apply up on migration 75 with status confirmed",
			setup: func(objID primitive.ObjectID) error {
				_, err := c.
					Database("test").
					Collection("users").
					InsertOne(ctx, map[string]interface{}{
						"_id":    objID,
						"status": models.UserStatusConfirmed.String(),
					})

				return err
			},
			expected: true,
		},
		{
			description: "Success to apply up on migration 75 with status unconfirmed",
			setup: func(objID primitive.ObjectID) error {
				_, err := c.
					Database("test").
					Collection("users").
					InsertOne(ctx, map[string]interface{}{
						"_id":    objID,
						"status": models.UserStatusNotConfirmed.String(),
					})

				return err
			},
			expected: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			tt.Cleanup(func() {
				assert.NoError(tt, srv.Reset())
			})

			objID := primitive.NewObjectID()
			assert.NoError(tt, tc.setup(objID))

			migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[74])
			require.NoError(tt, migrates.Up(context.Background(), migrate.AllAvailable))
			require.NoError(tt, migrates.Down(context.Background(), migrate.AllAvailable))

			query := c.
				Database("test").
				Collection("users").
				FindOne(context.TODO(), bson.M{"_id": objID})

			user := make(map[string]interface{})
			require.NoError(tt, query.Decode(&user))

			_, ok := user["status"]
			require.Equal(tt, false, ok)

			confirmed, ok := user["confirmed"].(bool)
			require.Equal(tt, true, ok)
			require.Equal(tt, tc.expected, confirmed)
		})
	}
}
