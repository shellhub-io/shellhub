package migrations

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
)

func TestMigration106Up(t *testing.T) {
	ctx := context.Background()

	cases := []struct {
		description string
		setup       func() error
		verify      func(tt *testing.T)
	}{
		{
			description: "succeeds creating indexes on devices collection",
			setup: func() error {
				devices := []bson.M{
					{
						"uid":               "device-1",
						"name":              "Device 1",
						"tenant_id":         "tenant-1",
						"status":            "removed",
						"status_updated_at": time.Now(),
					},
					{
						"uid":       "device-2",
						"name":      "Device 2",
						"tenant_id": "tenant-1",
						"status":    "accepted",
					},
				}

				_, err := c.Database("test").Collection("devices").InsertMany(ctx, []any{devices[0], devices[1]})

				return err
			},
			verify: func(tt *testing.T) {
				cursor, err := c.Database("test").Collection("devices").Indexes().List(ctx)
				require.NoError(tt, err)

				var indexes []bson.M
				require.NoError(tt, cursor.All(ctx, &indexes))

				indexNames := make([]string, 0, len(indexes))
				for _, index := range indexes {
					if name, ok := index["name"].(string); ok {
						indexNames = append(indexNames, name)
					}
				}

				assert.Contains(tt, indexNames, "idx_status_status_updated_at")
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			tt.Cleanup(func() { assert.NoError(tt, srv.Reset()) })

			require.NoError(tt, tc.setup())
			migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[105])
			require.NoError(tt, migrates.Up(ctx, migrate.AllAvailable))
			tc.verify(tt)
		})
	}
}

func TestMigration106Down(t *testing.T) {
	ctx := context.Background()

	t.Run("succeeds dropping created indexes", func(tt *testing.T) {
		tt.Cleanup(func() { assert.NoError(tt, srv.Reset()) })

		device := bson.M{
			"uid":       "device-1",
			"name":      "Device 1",
			"tenant_id": "tenant-1",
			"status":    "accepted",
		}
		_, err := c.Database("test").Collection("devices").InsertOne(ctx, device)
		require.NoError(tt, err)

		migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[105]) // migration106
		require.NoError(tt, migrates.Up(ctx, migrate.AllAvailable))
		require.NoError(tt, migrates.Down(ctx, migrate.AllAvailable))

		cursor, err := c.Database("test").Collection("devices").Indexes().List(ctx)
		require.NoError(tt, err)

		var indexes []bson.M
		require.NoError(tt, cursor.All(ctx, &indexes))

		indexNames := make([]string, 0, len(indexes))
		for _, index := range indexes {
			if name, ok := index["name"].(string); ok {
				indexNames = append(indexNames, name)
			}
		}

		assert.NotContains(tt, indexNames, "idx_status_status_updated_at")
		assert.Contains(tt, indexNames, "_id_")
	})
}
