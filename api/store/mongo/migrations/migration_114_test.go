package migrations

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
)

func TestMigration114Up(t *testing.T) {
	ctx := context.Background()

	cases := []struct {
		description string
		setup       func() error
		verify      func(tt *testing.T)
	}{
		{
			description: "succeeds adding admin=false to all users",
			setup: func() error {
				users := []bson.M{
					{"_id": "u1", "username": "user1", "email": "user1@example.com"},
					{"_id": "u2", "username": "user2", "email": "user2@example.com"},
				}

				_, err := c.Database("test").Collection("users").InsertMany(ctx, []any{users[0], users[1]})

				return err
			},

			verify: func(tt *testing.T) {
				cursor, err := c.Database("test").Collection("users").Find(ctx, bson.M{})
				require.NoError(tt, err)

				var results []bson.M
				require.NoError(tt, cursor.All(ctx, &results))

				for _, user := range results {
					val, exists := user["admin"]
					assert.True(tt, exists, "admin field should exist")
					assert.Equal(tt, false, val, "admin should be false")
				}
			},
		},

		{
			description: "succeeds when users collection is empty",
			setup: func() error {
				return nil
			},
			verify: func(tt *testing.T) {
				count, err := c.Database("test").Collection("users").CountDocuments(ctx, bson.M{})
				require.NoError(tt, err)

				assert.Equal(tt, int64(0), count, "users collection should be empty")
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			tt.Cleanup(func() { assert.NoError(tt, srv.Reset()) })
			require.NoError(tt, tc.setup())

			migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[113])
			require.NoError(tt, migrates.Up(ctx, migrate.AllAvailable))

			tc.verify(tt)
		})
	}
}

func TestMigration114Down(t *testing.T) {
	ctx := context.Background()

	cases := []struct {
		description string
		setup       func() error
		verify      func(tt *testing.T)
	}{
		{
			description: "succeeds removing admin field from all users",
			setup: func() error {
				users := []bson.M{
					{"_id": "u1", "username": "user1", "email": "user1@example.com", "admin": false},
					{"_id": "u2", "username": "user2", "email": "user2@example.com", "admin": true},
				}

				_, err := c.Database("test").Collection("users").InsertMany(ctx, []any{users[0], users[1]})

				return err
			},

			verify: func(tt *testing.T) {
				cursor, err := c.Database("test").Collection("users").Find(ctx, bson.M{})
				require.NoError(tt, err)

				var results []bson.M
				require.NoError(tt, cursor.All(ctx, &results))

				for _, user := range results {
					_, exists := user["admin"]
					assert.False(tt, exists, "admin field should not exist")
				}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			tt.Cleanup(func() { assert.NoError(tt, srv.Reset()) })
			require.NoError(tt, tc.setup())

			migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[113])
			require.NoError(tt, migrates.Up(ctx, migrate.AllAvailable))
			require.NoError(tt, migrates.Down(ctx, migrate.AllAvailable))

			tc.verify(tt)
		})
	}
}
