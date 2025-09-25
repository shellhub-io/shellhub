package migrations

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
)

func TestMigration111Up(t *testing.T) {
	ctx := context.Background()

	cases := []struct {
		description string
		setup       func() error
		verify      func(tt *testing.T)
	}{
		{
			description: "succeeds adding removed_at=null to all devices",
			setup: func() error {
				devices := []bson.M{{"_id": "d1", "status": "active"}, {"_id": "d2", "status": "inactive"}}
				_, err := c.Database("test").Collection("devices").InsertMany(ctx, []any{devices[0], devices[1]})

				return err
			},
			verify: func(tt *testing.T) {
				cursor, err := c.Database("test").Collection("devices").Find(ctx, bson.M{})
				require.NoError(tt, err)

				var results []bson.M
				require.NoError(tt, cursor.All(ctx, &results))

				for _, d := range results {
					val, exists := d["removed_at"]
					assert.True(tt, exists)
					assert.Nil(tt, val)
				}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			tt.Cleanup(func() { assert.NoError(tt, srv.Reset()) })

			require.NoError(tt, tc.setup())
			migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[110]) // migration111
			require.NoError(tt, migrates.Up(ctx, migrate.AllAvailable))
			tc.verify(tt)
		})
	}
}

func TestMigration111Down(t *testing.T) {
	ctx := context.Background()

	cases := []struct {
		description string
		setup       func() error
		verify      func(tt *testing.T)
	}{
		{
			description: "succeeds removing removed_at field from all devices",
			setup: func() error {
				devices := []bson.M{{"_id": "d1", "status": "active", "removed_at": nil}, {"_id": "d2", "status": "inactive", "removed_at": nil}}
				_, err := c.Database("test").Collection("devices").InsertMany(ctx, []any{devices[0], devices[1]})

				return err
			},
			verify: func(tt *testing.T) {
				cursor, err := c.Database("test").Collection("devices").Find(ctx, bson.M{})
				require.NoError(tt, err)

				var results []bson.M
				require.NoError(tt, cursor.All(ctx, &results))

				for _, d := range results {
					_, exists := d["removed_at"]
					assert.False(tt, exists)
				}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			tt.Cleanup(func() { assert.NoError(tt, srv.Reset()) })

			require.NoError(tt, tc.setup())
			migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[110])
			require.NoError(tt, migrates.Up(ctx, migrate.AllAvailable))
			require.NoError(tt, migrates.Down(ctx, migrate.AllAvailable))
			tc.verify(tt)
		})
	}
}
