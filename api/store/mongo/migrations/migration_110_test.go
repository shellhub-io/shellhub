package migrations

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
)

func TestMigration110Up(t *testing.T) {
	ctx := context.Background()

	cases := []struct {
		description string
		setup       func() error
		verify      func(tt *testing.T)
	}{
		{
			description: "removes all devices with status=removed",
			setup: func() error {
				devices := []bson.M{
					{"_id": "d1", "status": "accepted"},
					{"_id": "d2", "status": "removed"},
					{"_id": "d3", "status": "removed"},
				}

				_, err := c.Database("test").Collection("devices").InsertMany(ctx, []any{devices[0], devices[1], devices[2]})

				return err
			},
			verify: func(tt *testing.T) {
				cursor, err := c.Database("test").Collection("devices").Find(ctx, bson.M{})
				require.NoError(tt, err)

				var result []bson.M
				require.NoError(tt, cursor.All(ctx, &result))

				assert.Len(tt, result, 1)
				assert.Equal(tt, "accepted", result[0]["status"])
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			tt.Cleanup(func() { assert.NoError(tt, srv.Reset()) })

			require.NoError(tt, tc.setup())
			migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[109])
			require.NoError(tt, migrates.Up(ctx, migrate.AllAvailable))
			tc.verify(tt)
		})
	}
}
