package migrations

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestMigration105Up(t *testing.T) {
	ctx := context.Background()

	cases := []struct {
		description string
		setup       func() error
		verify      func(tt *testing.T)
	}{
		{
			description: "succeeds dropping removed_devices collection",
			setup: func() error {
				removedDevice := bson.M{
					"device": bson.M{
						"_id":       primitive.NewObjectID(),
						"uid":       "device-1",
						"name":      "Device 1",
						"tenant_id": "tenant-1",
					},
					"timestamp": time.Now(),
				}

				_, err := c.Database("test").Collection("removed_devices").InsertOne(ctx, removedDevice)

				return err
			},
			verify: func(tt *testing.T) {
				collections, err := c.Database("test").ListCollectionNames(ctx, bson.M{"name": "removed_devices"})
				require.NoError(tt, err)
				assert.Empty(tt, collections)
			},
		},
		{
			description: "handles non-existent collection gracefully",
			setup: func() error {
				return nil
			},
			verify: func(tt *testing.T) {
				collections, err := c.Database("test").ListCollectionNames(ctx, bson.M{"name": "removed_devices"})
				require.NoError(tt, err)
				assert.Empty(tt, collections)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			tt.Cleanup(func() { assert.NoError(tt, srv.Reset()) })

			require.NoError(tt, tc.setup())
			migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[104])
			require.NoError(tt, migrates.Up(ctx, migrate.AllAvailable))
			tc.verify(tt)
		})
	}
}
