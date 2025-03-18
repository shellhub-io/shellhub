package migrations

import (
	"context"
	"slices"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestMigration94Up(t *testing.T) {
	cases := []struct {
		description string
		setup       func(ctx context.Context) error
		verify      func(ctx context.Context, tt *testing.T)
	}{
		{
			description: "drops the 'connected_devices' collection",
			setup: func(ctx context.Context) error {
				_, err := c.Database("test").Collection("connected_devices").InsertOne(ctx, bson.M{"uid": "auth-test"})

				return err
			},
			verify: func(ctx context.Context, tt *testing.T) {
				res, err := c.Database("test").ListCollectionNames(ctx, bson.M{})
				require.NoError(tt, err)
				require.Equal(tt, false, slices.Contains(res, "connected_devices"))
			},
		},
		{
			description: "sets the value to nil when the device have a related connected_device",
			setup: func(ctx context.Context) error {
				randomUIDs := []string{uuid.Generate(), uuid.Generate()}
				for _, uid := range randomUIDs {
					if _, err := c.Database("test").Collection("devices").InsertOne(ctx, bson.M{"uid": uid, "last_seen": time.Now()}); err != nil {
						return err
					}

					if _, err := c.Database("test").Collection("connected_devices").InsertOne(ctx, bson.M{"uid": uid}); err != nil {
						return err
					}
				}

				return nil
			},
			verify: func(ctx context.Context, tt *testing.T) {
				cursor, err := c.Database("test").Collection("devices").Find(ctx, bson.M{})
				require.NoError(tt, err)
				defer cursor.Close(ctx)

				for cursor.Next(ctx) {
					device := make(map[string]any)
					require.NoError(tt, cursor.Decode(&device))

					disconnectedAt, ok := device["disconnected_at"]
					require.Equal(tt, true, ok)
					require.Equal(tt, nil, disconnectedAt)
				}
			},
		},
		{
			description: "sets the value to last_seen when the device does not have a related connected_device",
			setup: func(ctx context.Context) error {
				randomUIDs := []string{uuid.Generate(), uuid.Generate()}
				for _, uid := range randomUIDs {
					_, err := c.Database("test").Collection("devices").InsertOne(ctx, bson.M{"uid": uid, "last_seen": time.Now()})

					return err
				}

				return nil
			},
			verify: func(ctx context.Context, tt *testing.T) {
				cursor, err := c.Database("test").Collection("devices").Find(ctx, bson.M{})
				require.NoError(tt, err)
				defer cursor.Close(ctx)

				for cursor.Next(ctx) {
					device := make(map[string]any)
					require.NoError(tt, cursor.Decode(&device))

					disconnectedAt, ok := device["disconnected_at"]
					require.Equal(tt, true, ok)
					require.WithinDuration(tt, device["last_seen"].(primitive.DateTime).Time(), disconnectedAt.(primitive.DateTime).Time(), 1*time.Second)
				}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			ctx := context.Background()

			tt.Cleanup(func() {
				assert.NoError(tt, srv.Reset())
			})

			require.NoError(tt, tc.setup(ctx))

			migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[93])
			require.NoError(tt, migrates.Up(ctx, migrate.AllAvailable))

			tc.verify(ctx, tt)
		})
	}
}
