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
)

func TestMigration83Up(t *testing.T) {
	ctx := context.Background()

	mock := &envmock.Backend{}
	envs.DefaultBackend = mock

	cases := []struct {
		description string
		setup       func() error
		test        func() error
	}{
		{
			description: "Success to apply up on migration 83",
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
		t.Run(tc.description, func(tt *testing.T) {
			tt.Cleanup(func() {
				assert.NoError(tt, srv.Reset())
			})

			assert.NoError(tt, tc.setup())

			migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[82])
			require.NoError(tt, migrates.Up(context.Background(), migrate.AllAvailable))

			query := c.
				Database("test").
				Collection("users").
				FindOne(context.TODO(), bson.M{"name": "john doe"})

			user := make(map[string]interface{})
			require.NoError(tt, query.Decode(&user))

			v, ok := user["origin"]
			require.Equal(tt, true, ok)
			require.Equal(tt, v, models.UserOriginLocal.String())
		})
	}
}

func TestMigration83Down(t *testing.T) {
	ctx := context.Background()

	mock := &envmock.Backend{}
	envs.DefaultBackend = mock

	cases := []struct {
		description string
		setup       func() error
		test        func() error
	}{
		{
			description: "Success to apply up on migration 83",
			setup: func() error {
				_, err := c.
					Database("test").
					Collection("users").
					InsertOne(ctx, map[string]interface{}{
						"name":   "john doe",
						"origin": models.UserOriginLocal.String(),
					})

				return err
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			t.Cleanup(func() {
				assert.NoError(t, srv.Reset())
			})

			assert.NoError(t, tc.setup())

			migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[82])
			require.NoError(t, migrates.Up(context.Background(), migrate.AllAvailable))
			require.NoError(t, migrates.Down(context.Background(), migrate.AllAvailable))

			query := c.
				Database("test").
				Collection("users").
				FindOne(context.TODO(), bson.M{"name": "john doe"})

			user := make(map[string]interface{})
			require.NoError(t, query.Decode(&user))

			_, ok := user["origin"]
			require.Equal(t, false, ok)
		})
	}
}
