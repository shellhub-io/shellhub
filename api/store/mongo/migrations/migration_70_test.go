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

func TestMigration70Up(t *testing.T) {
	ctx := context.Background()

	mock := &envmock.Backend{}
	envs.DefaultBackend = mock

	cases := []struct {
		description string
		setup       func() error
		test        func() error
	}{
		{
			description: "Success to apply up on migration 70",
			setup: func() error {
				_, err := c.
					Database("test").
					Collection("users").
					InsertOne(ctx, map[string]interface{}{
						"name": "john doe",
					})

				return err
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

			migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[69])
			require.NoError(t, migrates.Up(context.Background(), migrate.AllAvailable))

			query := c.
				Database("test").
				Collection("users").
				FindOne(context.TODO(), bson.M{"name": "john doe"})

			user := make(map[string]interface{})
			require.NoError(t, query.Decode(&user))

			_, ok := user["preferences"]
			require.Equal(t, true, ok)
		})
	}
}

func TestMigration70Down(t *testing.T) {
	ctx := context.Background()

	mock := &envmock.Backend{}
	envs.DefaultBackend = mock

	cases := []struct {
		description string
		setup       func() error
		test        func() error
	}{
		{
			description: "Success to apply up on migration 70",
			setup: func() error {
				_, err := c.
					Database("test").
					Collection("users").
					InsertOne(ctx, map[string]interface{}{
						"name":        "john doe",
						"preferences": map[string]interface{}{},
					})

				return err
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

			migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[69])
			require.NoError(t, migrates.Up(context.Background(), migrate.AllAvailable))
			require.NoError(t, migrates.Down(context.Background(), migrate.AllAvailable))

			query := c.
				Database("test").
				Collection("users").
				FindOne(context.TODO(), bson.M{"name": "john doe"})

			user := make(map[string]interface{})
			require.NoError(t, query.Decode(&user))

			_, ok := user["preferences"]
			require.Equal(t, false, ok)
		})
	}
}
