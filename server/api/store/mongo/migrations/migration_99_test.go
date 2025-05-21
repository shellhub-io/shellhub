package migrations

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
)

func TestMigration99Up(t *testing.T) {
	ctx := context.Background()

	cases := []struct {
		description string
		setup       func() error
		verify      func(tt *testing.T)
	}{
		{
			description: "succeeds",
			setup: func() error {
				_, err := c.
					Database("test").
					Collection("users").
					InsertOne(ctx, bson.M{"email": "john.doe@test.com", "username": ""})

				return err
			},
			verify: func(tt *testing.T) {
				user := make(map[string]any)
				require.NoError(tt, c.Database("test").Collection("users").FindOne(ctx, bson.M{"email": "john.doe@test.com"}).Decode(&user))

				username, ok := user["username"]
				require.Equal(tt, true, ok)
				require.Equal(tt, nil, username)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			tt.Cleanup(func() {
				assert.NoError(tt, srv.Reset())
			})

			require.NoError(tt, tc.setup())

			migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[98])
			require.NoError(tt, migrates.Up(ctx, migrate.AllAvailable))

			tc.verify(tt)
		})
	}
}

func TestMigration99Down(t *testing.T) {
	ctx := context.Background()

	cases := []struct {
		description string
		setup       func() error
		verify      func(tt *testing.T)
	}{
		{
			description: "succeeds",
			setup: func() error {
				_, err := c.
					Database("test").
					Collection("users").
					InsertOne(ctx, bson.M{"email": "john.doe@test.com", "username": nil})

				return err
			},
			verify: func(tt *testing.T) {
				user := make(map[string]any)
				require.NoError(tt, c.Database("test").Collection("users").FindOne(ctx, bson.M{"email": "john.doe@test.com"}).Decode(&user))

				username, ok := user["username"]
				require.Equal(tt, true, ok)
				require.Equal(tt, "", username)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			tt.Cleanup(func() {
				assert.NoError(tt, srv.Reset())
			})

			require.NoError(tt, tc.setup())

			migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[98])
			require.NoError(tt, migrates.Up(ctx, migrate.AllAvailable))
			require.NoError(tt, migrates.Down(ctx, migrate.AllAvailable))

			tc.verify(tt)
		})
	}
}
