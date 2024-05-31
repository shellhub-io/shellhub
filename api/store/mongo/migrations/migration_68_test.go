package migrations

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/envs"
	envMocks "github.com/shellhub-io/shellhub/pkg/envs/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
)

func TestMigration68Up(t *testing.T) {
	ctx := context.Background()

	mock := &envMocks.Backend{}
	envs.DefaultBackend = mock

	cases := []struct {
		description string
		setup       func() error
		test        func() error
		expected    map[string]interface{}
	}{
		{
			description: "Success to apply up on migration 68",
			setup: func() error {
				_, err := c.
					Database("test").
					Collection("api_keys").
					InsertOne(ctx, map[string]interface{}{
						"name":    "dev",
						"user_id": "000000000000000000000000",
					})

				return err
			},
			expected: map[string]interface{}{
				"name":       "dev",
				"created_by": "000000000000000000000000",
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.description, func(t *testing.T) {
			t.Cleanup(func() {
				assert.NoError(t, srv.Reset())
			})

			assert.NoError(t, tc.setup())

			migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[67])
			require.NoError(t, migrates.Up(context.Background(), migrate.AllAvailable))

			query := c.
				Database("test").
				Collection("api_keys").
				FindOne(context.TODO(), bson.M{"name": "dev"})

			apiKey := make(map[string]interface{})
			require.NoError(t, query.Decode(&apiKey))

			_, ok := apiKey["user_id"]
			require.Equal(t, false, ok)

			attr, ok := apiKey["created_by"]
			require.Equal(t, true, ok)
			require.Equal(t, tc.expected["created_by"], attr)
		})
	}
}

func TestMigration68Down(t *testing.T) {
	ctx := context.Background()

	mock := &envMocks.Backend{}
	envs.DefaultBackend = mock

	cases := []struct {
		description string
		setup       func() error
		test        func() error
		expected    map[string]interface{}
	}{
		{
			description: "Success to apply down on migration 68",
			setup: func() error {
				_, err := c.
					Database("test").
					Collection("api_keys").
					InsertOne(ctx, map[string]interface{}{
						"name":       "dev",
						"created_by": "000000000000000000000000",
					})

				return err
			},
			expected: map[string]interface{}{
				"name":    "dev",
				"user_id": "000000000000000000000000",
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.description, func(t *testing.T) {
			t.Cleanup(func() {
				assert.NoError(t, srv.Reset())
			})

			assert.NoError(t, tc.setup())

			migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[67])
			require.NoError(t, migrates.Up(context.Background(), migrate.AllAvailable))
			require.NoError(t, migrates.Down(context.Background(), migrate.AllAvailable))

			query := c.
				Database("test").
				Collection("api_keys").
				FindOne(context.TODO(), bson.M{"name": "dev"})

			apiKey := make(map[string]interface{})
			require.NoError(t, query.Decode(&apiKey))

			_, ok := apiKey["created_by"]
			require.Equal(t, false, ok)

			attr, ok := apiKey["user_id"]
			require.Equal(t, true, ok)
			require.Equal(t, tc.expected["user_id"], attr)
		})
	}
}
