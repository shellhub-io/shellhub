package migrations

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
)

func TestMigration109Up(t *testing.T) {
	ctx := context.Background()

	cases := []struct {
		description string
		setup       func() error
		verify      func(tt *testing.T)
	}{
		{
			description: "succeeds creating indexes on tags collection",
			setup: func() error {
				tags := []bson.M{{"tenant_id": "tenant-1", "name": "tag1"}, {"tenant_id": "tenant-1", "name": "tag2"}}
				_, err := c.Database("test").Collection("tags").InsertMany(ctx, []any{tags[0], tags[1]})

				return err
			},
			verify: func(tt *testing.T) {
				cursor, err := c.Database("test").Collection("tags").Indexes().List(ctx)
				require.NoError(tt, err)

				var indexes []bson.M
				require.NoError(tt, cursor.All(ctx, &indexes))

				indexNames := make([]string, 0, len(indexes))
				for _, index := range indexes {
					if name, ok := index["name"].(string); ok {
						indexNames = append(indexNames, name)
					}
				}

				assert.Contains(tt, indexNames, "idx_tenant_id_name_unique")
				assert.Contains(tt, indexNames, "idx_tenant_id")
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			tt.Cleanup(func() { assert.NoError(tt, srv.Reset()) })

			require.NoError(tt, tc.setup())
			migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[108])
			require.NoError(tt, migrates.Up(ctx, migrate.AllAvailable))
			tc.verify(tt)
		})
	}
}

func TestMigration109Down(t *testing.T) {
	ctx := context.Background()

	t.Run("succeeds dropping created indexes", func(tt *testing.T) {
		tt.Cleanup(func() { assert.NoError(tt, srv.Reset()) })

		tag := bson.M{"tenant_id": "tenant-1", "name": "tag1"}
		_, err := c.Database("test").Collection("tags").InsertOne(ctx, tag)
		require.NoError(tt, err)

		migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[108]) // migration109
		require.NoError(tt, migrates.Up(ctx, migrate.AllAvailable))
		require.NoError(tt, migrates.Down(ctx, migrate.AllAvailable))

		cursor, err := c.Database("test").Collection("tags").Indexes().List(ctx)
		require.NoError(tt, err)

		var indexes []bson.M
		require.NoError(tt, cursor.All(ctx, &indexes))

		indexNames := make([]string, 0, len(indexes))
		for _, index := range indexes {
			if name, ok := index["name"].(string); ok {
				indexNames = append(indexNames, name)
			}
		}

		assert.NotContains(tt, indexNames, "idx_tenant_id_name_unique")
		assert.NotContains(tt, indexNames, "idx_tenant_id")
		assert.Contains(tt, indexNames, "_id_")
	})
}
