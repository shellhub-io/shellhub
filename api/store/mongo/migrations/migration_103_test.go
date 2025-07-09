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

func TestMigration103Up(t *testing.T) {
	ctx := context.Background()

	cases := []struct {
		description string
		setup       func() error
		verify      func(tt *testing.T)
	}{
		{
			description: "succeeds converting devices_removed to devices with status removed",
			setup: func() error {
				timestamp := time.Now()

				removedDevices := []bson.M{
					{
						"device": bson.M{
							"_id":        primitive.NewObjectID(),
							"uid":        "device-1",
							"name":       "Device 1",
							"tenant_id":  "tenant-1",
							"status":     "accepted",
							"created_at": timestamp,
						},
						"timestamp": timestamp,
					},
					{
						"device": bson.M{
							"_id":        primitive.NewObjectID(),
							"uid":        "device-2",
							"name":       "Device 2",
							"tenant_id":  "tenant-2",
							"status":     "pending",
							"created_at": timestamp,
						},
						"timestamp": timestamp,
					},
				}

				existingID := primitive.NewObjectID()
				existingDevice := bson.M{
					"_id":        existingID,
					"uid":        "device-3",
					"name":       "Existing Device",
					"tenant_id":  "tenant-1",
					"status":     "accepted",
					"created_at": timestamp,
				}

				removedDeviceExisting := bson.M{"device": existingDevice, "timestamp": timestamp}

				removedDeviceNil := bson.M{"device": nil, "timestamp": timestamp}

				if _, err := c.Database("test").Collection("devices").InsertOne(ctx, existingDevice); err != nil {
					return err
				}

				allRemovedDevices := []any{
					removedDevices[0],
					removedDevices[1],
					removedDeviceExisting,
					removedDeviceNil,
				}

				_, err := c.Database("test").Collection("removed_devices").InsertMany(ctx, allRemovedDevices)

				return err
			},
			verify: func(tt *testing.T) {
				var device1 map[string]any
				require.NoError(tt, c.Database("test").Collection("devices").FindOne(ctx, bson.M{"uid": "device-1"}).Decode(&device1))
				assert.Equal(tt, "device-1", device1["uid"])
				assert.Equal(tt, "Device 1", device1["name"])
				assert.Equal(tt, "tenant-1", device1["tenant_id"])
				assert.Equal(tt, "removed", device1["status"])
				assert.NotNil(tt, device1["status_updated_at"])

				var device2 map[string]any
				require.NoError(tt, c.Database("test").Collection("devices").FindOne(ctx, bson.M{"uid": "device-2"}).Decode(&device2))
				assert.Equal(tt, "device-2", device2["uid"])
				assert.Equal(tt, "Device 2", device2["name"])
				assert.Equal(tt, "tenant-2", device2["tenant_id"])
				assert.Equal(tt, "removed", device2["status"])
				assert.NotNil(tt, device2["status_updated_at"])

				var device3 map[string]any
				require.NoError(tt, c.Database("test").Collection("devices").FindOne(ctx, bson.M{"uid": "device-3"}).Decode(&device3))
				assert.Equal(tt, "device-3", device3["uid"])
				assert.Equal(tt, "Existing Device", device3["name"])
				assert.Equal(tt, "accepted", device3["status"])

				count, err := c.Database("test").Collection("devices").CountDocuments(ctx, bson.M{})
				require.NoError(tt, err)
				assert.Equal(tt, int64(3), count)

				removedCount, err := c.Database("test").Collection("devices").CountDocuments(ctx, bson.M{"status": "removed"})
				require.NoError(tt, err)
				assert.Equal(tt, int64(2), removedCount)
			},
		},
		{
			description: "handles empty removed_devices collection",
			setup: func() error {
				return nil
			},
			verify: func(tt *testing.T) {
				count, err := c.Database("test").Collection("devices").CountDocuments(ctx, bson.M{})
				require.NoError(tt, err)
				assert.Equal(tt, int64(0), count)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			tt.Cleanup(func() { assert.NoError(tt, srv.Reset()) })

			require.NoError(tt, tc.setup())
			migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[102])
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
			description: "succeeds reverting devices with status removed back to removed_devices",
			setup: func() error {
				timestamp := time.Now()

				devices := []bson.M{
					{
						"_id":               primitive.NewObjectID(),
						"uid":               "device-1",
						"name":              "Device 1",
						"tenant_id":         "tenant-1",
						"public_key":        "key1",
						"status":            "removed",
						"status_updated_at": timestamp,
						"created_at":        timestamp,
					},
					{
						"_id":               primitive.NewObjectID(),
						"uid":               "device-2",
						"name":              "Device 2",
						"tenant_id":         "tenant-2",
						"public_key":        "key2",
						"status":            "removed",
						"status_updated_at": timestamp,
						"created_at":        timestamp,
					},
					{
						"_id":        primitive.NewObjectID(),
						"uid":        "device-3",
						"name":       "Device 3",
						"tenant_id":  "tenant-1",
						"status":     "accepted",
						"created_at": timestamp,
					},
					{
						"_id":        primitive.NewObjectID(),
						"uid":        "device-4",
						"name":       "Device 4",
						"tenant_id":  "tenant-1",
						"status":     "removed",
						"created_at": timestamp,
					},
				}

				_, err := c.Database("test").Collection("devices").InsertMany(ctx, []any{devices[0], devices[1], devices[2], devices[3]})

				return err
			},
			verify: func(tt *testing.T) {
				removedCount, err := c.Database("test").Collection("devices").CountDocuments(ctx, bson.M{"status": "removed"})
				require.NoError(tt, err)
				assert.Equal(tt, int64(0), removedCount)

				totalCount, err := c.Database("test").Collection("devices").CountDocuments(ctx, bson.M{})
				require.NoError(tt, err)
				assert.Equal(tt, int64(1), totalCount)

				var device map[string]any
				require.NoError(tt, c.Database("test").Collection("devices").FindOne(ctx, bson.M{"uid": "device-3"}).Decode(&device))
				assert.Equal(tt, "device-3", device["uid"])
				assert.Equal(tt, "accepted", device["status"])

				removedDevicesCount, err := c.Database("test").Collection("removed_devices").CountDocuments(ctx, bson.M{})
				require.NoError(tt, err)
				assert.Equal(tt, int64(2), removedDevicesCount)

				var removedDevice1 map[string]any
				require.NoError(tt, c.Database("test").Collection("removed_devices").FindOne(ctx, bson.M{"device.uid": "device-1"}).Decode(&removedDevice1))

				device1Data := removedDevice1["device"].(map[string]any)
				assert.Equal(tt, "device-1", device1Data["uid"])
				assert.Equal(tt, "Device 1", device1Data["name"])
				assert.Equal(tt, "tenant-1", device1Data["tenant_id"])
				assert.Equal(tt, "key1", device1Data["public_key"])

				assert.NotNil(tt, removedDevice1["timestamp"])

				var removedDevice2 map[string]any
				require.NoError(tt, c.Database("test").Collection("removed_devices").FindOne(ctx, bson.M{"device.uid": "device-2"}).Decode(&removedDevice2))

				device2Data := removedDevice2["device"].(map[string]any)
				assert.Equal(tt, "device-2", device2Data["uid"])
				assert.Equal(tt, "Device 2", device2Data["name"])
				assert.Equal(tt, "tenant-2", device2Data["tenant_id"])
				assert.Equal(tt, "key2", device2Data["public_key"])

				assert.NotNil(tt, removedDevice2["timestamp"])

				device4Count, err := c.Database("test").Collection("removed_devices").CountDocuments(ctx, bson.M{"device.uid": "device-4"})
				require.NoError(tt, err)
				assert.Equal(tt, int64(0), device4Count)
			},
		},
		{
			description: "handles empty devices collection with status removed",
			setup: func() error {
				timestamp := time.Now()

				devices := []bson.M{
					{
						"_id":        primitive.NewObjectID(),
						"uid":        "device-1",
						"name":       "Device 1",
						"tenant_id":  "tenant-1",
						"status":     "accepted",
						"created_at": timestamp,
					},
				}

				_, err := c.Database("test").Collection("devices").InsertMany(ctx, []any{devices[0]})

				return err
			},
			verify: func(tt *testing.T) {
				totalCount, err := c.Database("test").Collection("devices").CountDocuments(ctx, bson.M{})
				require.NoError(tt, err)
				assert.Equal(tt, int64(1), totalCount)

				removedCount, err := c.Database("test").Collection("removed_devices").CountDocuments(ctx, bson.M{})
				require.NoError(tt, err)
				assert.Equal(tt, int64(0), removedCount)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			tt.Cleanup(func() { assert.NoError(tt, srv.Reset()) })

			require.NoError(tt, tc.setup())
			migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[102])
			require.NoError(tt, migrates.Up(ctx, migrate.AllAvailable))
			require.NoError(tt, migrates.Down(ctx, migrate.AllAvailable))
			tc.verify(tt)
		})
	}
}
