package migrations

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/envs"
	envmock "github.com/shellhub-io/shellhub/pkg/envs/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
)

func TestMigration87Up(t *testing.T) {
	ctx := context.Background()

	mock := &envmock.Backend{}
	envs.DefaultBackend = mock

	mock.On("Get", "SHELLHUB_CLOUD").Return("false")
	mock.On("Get", "SHELLHUB_ENTERPRISE").Return("false")

	tests := []struct {
		description string
		setup       func() error
	}{
		{
			description: "Apply up on migration 87",
			setup: func() error {
				_, err := c.
					Database("test").
					Collection("system").
					InsertOne(ctx, map[string]interface{}{})

				return err
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(tt *testing.T) {
			tt.Cleanup(func() {
				assert.NoError(tt, srv.Reset())
			})

			assert.NoError(tt, tc.setup())

			migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[86])
			require.NoError(tt, migrates.Up(context.Background(), migrate.AllAvailable))

			query := c.
				Database("test").
				Collection("system").
				FindOne(context.TODO(), bson.M{})

			system := make(map[string]interface{})
			require.NoError(tt, query.Decode(&system))

			authentication, ok := system["authentication"].(map[string]interface{})
			require.Equal(tt, true, ok)
			require.Equal(tt, map[string]interface{}{"local": map[string]interface{}{"enabled": true}}, authentication)
		})
	}
}

func TestMigration87Down(t *testing.T) {
	ctx := context.Background()

	mock := &envmock.Backend{}
	envs.DefaultBackend = mock

	mock.On("Get", "SHELLHUB_CLOUD").Return("false")
	mock.On("Get", "SHELLHUB_ENTERPRISE").Return("false")

	tests := []struct {
		description string
		setup       func() error
	}{
		{
			description: "Apply up on migration 87",
			setup: func() error {
				_, err := c.
					Database("test").
					Collection("system").
					InsertOne(ctx, map[string]interface{}{
						"authentication": "some_value",
					})

				return err
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(tt *testing.T) {
			tt.Cleanup(func() {
				assert.NoError(tt, srv.Reset())
			})

			assert.NoError(tt, tc.setup())

			migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[86])
			require.NoError(tt, migrates.Up(context.Background(), migrate.AllAvailable))
			require.NoError(tt, migrates.Down(context.Background(), migrate.AllAvailable))

			query := c.
				Database("test").
				Collection("system").
				FindOne(context.TODO(), bson.M{})

			system := make(map[string]interface{})
			require.NoError(tt, query.Decode(&system))

			_, ok := system["authentication"]
			require.Equal(tt, false, ok)
		})
	}
}
