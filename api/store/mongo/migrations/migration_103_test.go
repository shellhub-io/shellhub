package migrations

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
)

func TestMigration103Up(t *testing.T) {
	ctx := context.Background()

	cases := []struct {
		description string
		setup       func() error
		verify      func(tt *testing.T)
	}{
		{
			description: "succeeds with namespaces and removed devices",
			setup: func() error {
				namespaces := []bson.M{
					{
						"tenant_id": "tenant-1",
						"name":      "namespace-1",
					},
					{
						"tenant_id": "tenant-2",
						"name":      "namespace-2",
					},
					{
						"tenant_id": "tenant-3",
						"name":      "namespace-3",
					},
				}

				_, err := c.Database("test").Collection("namespaces").InsertMany(ctx, []any{namespaces[0], namespaces[1], namespaces[2]})
				if err != nil {
					return err
				}

				removedDevices := []bson.M{
					// tenant-1: 5 removed devices
					{"device": bson.M{"tenant_id": "tenant-1"}, "timestamp": "2024-01-01"},
					{"device": bson.M{"tenant_id": "tenant-1"}, "timestamp": "2024-01-02"},
					{"device": bson.M{"tenant_id": "tenant-1"}, "timestamp": "2024-01-03"},
					{"device": bson.M{"tenant_id": "tenant-1"}, "timestamp": "2024-01-04"},
					{"device": bson.M{"tenant_id": "tenant-1"}, "timestamp": "2024-01-05"},

					// tenant-2: 2 removed devices
					{"device": bson.M{"tenant_id": "tenant-2"}, "timestamp": "2024-01-06"},
					{"device": bson.M{"tenant_id": "tenant-2"}, "timestamp": "2024-01-07"},

					// tenant-3: no removed devices (test empty case)
				}

				_, err = c.
					Database("test").
					Collection("removed_devices").
					InsertMany(ctx, []any{removedDevices[0], removedDevices[1], removedDevices[2], removedDevices[3], removedDevices[4], removedDevices[5], removedDevices[6]})

				return err
			},
			verify: func(tt *testing.T) {
				namespace1 := make(map[string]any)
				require.NoError(tt, c.Database("test").Collection("namespaces").FindOne(ctx, bson.M{"tenant_id": "tenant-1"}).Decode(&namespace1))
				assert.Equal(tt, int32(5), namespace1["devices_removed_count"].(int32))

				namespace2 := make(map[string]any)
				require.NoError(tt, c.Database("test").Collection("namespaces").FindOne(ctx, bson.M{"tenant_id": "tenant-2"}).Decode(&namespace2))
				assert.Equal(tt, int32(2), namespace2["devices_removed_count"].(int32))

				namespace3 := make(map[string]any)
				require.NoError(tt, c.Database("test").Collection("namespaces").FindOne(ctx, bson.M{"tenant_id": "tenant-3"}).Decode(&namespace3))
				assert.Equal(tt, int32(0), namespace3["devices_removed_count"])
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			tt.Cleanup(func() { assert.NoError(tt, srv.Reset()) })
			require.NoError(tt, tc.setup())

			migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[101])
			require.NoError(tt, migrates.Up(ctx, migrate.AllAvailable))
			tc.verify(tt)
		})
	}
}

func TestMigration103Down(t *testing.T) {
	ctx := context.Background()

	cases := []struct {
		description string
		setup       func() error
		verify      func(tt *testing.T)
	}{
		{
			description: "succeeds",
			setup: func() error {
				namespaces := []bson.M{
					{
						"tenant_id":             "tenant-1",
						"name":                  "namespace-1",
						"devices_removed_count": 10,
					},
					{
						"tenant_id":             "tenant-2",
						"name":                  "namespace-2",
						"devices_removed_count": 5,
					},
				}
				_, err := c.Database("test").Collection("namespaces").InsertMany(ctx, []any{namespaces[0], namespaces[1]})

				return err
			},
			verify: func(tt *testing.T) {
				namespace1 := make(map[string]any)
				require.NoError(tt, c.Database("test").Collection("namespaces").FindOne(ctx, bson.M{"tenant_id": "tenant-1"}).Decode(&namespace1))

				_, hasRemovedCount1 := namespace1["devices_removed_count"]
				assert.False(tt, hasRemovedCount1)
				assert.Equal(tt, "namespace-1", namespace1["name"])

				namespace2 := make(map[string]any)
				require.NoError(tt, c.Database("test").Collection("namespaces").FindOne(ctx, bson.M{"tenant_id": "tenant-2"}).Decode(&namespace2))

				_, hasRemovedCount2 := namespace2["devices_removed_count"]
				assert.False(tt, hasRemovedCount2)
				assert.Equal(tt, "namespace-2", namespace2["name"])
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			tt.Cleanup(func() {
				assert.NoError(tt, srv.Reset())
			})

			require.NoError(tt, tc.setup())

			migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[102])
			require.NoError(tt, migrates.Up(ctx, migrate.AllAvailable))
			require.NoError(tt, migrates.Down(ctx, migrate.AllAvailable))

			tc.verify(tt)
		})
	}
}
