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

func TestMigration104Up(t *testing.T) {
	ctx := context.Background()

	cases := []struct {
		description string
		setup       func() error
		verify      func(tt *testing.T)
	}{
		{
			description: "succeeds adding devices_removed_count field based on existing removed devices",
			setup: func() error {
				timestamp := time.Now()

				namespaces := []bson.M{
					{
						"tenant_id":              "tenant-1",
						"devices_accepted_count": 5,
						"devices_pending_count":  2,
						"devices_rejected_count": 1,
					},
					{
						"tenant_id":              "tenant-2",
						"devices_accepted_count": 3,
						"devices_pending_count":  0,
						"devices_rejected_count": 0,
					},
					{
						"tenant_id":              "tenant-3",
						"devices_accepted_count": 0,
						"devices_pending_count":  0,
						"devices_rejected_count": 0,
					},
				}

				devices := []bson.M{
					// tenant-1: 3 removed device
					{
						"uid":               "device-1",
						"name":              "Device 1",
						"tenant_id":         "tenant-1",
						"status":            "removed",
						"status_updated_at": timestamp,
						"created_at":        timestamp,
					},
					{
						"uid":               "device-2",
						"name":              "Device 2",
						"tenant_id":         "tenant-1",
						"status":            "removed",
						"status_updated_at": timestamp,
						"created_at":        timestamp,
					},
					{
						"uid":               "device-3",
						"name":              "Device 3",
						"tenant_id":         "tenant-1",
						"status":            "removed",
						"status_updated_at": timestamp,
						"created_at":        timestamp,
					},
					// tenant-1: non-removed devices
					{
						"uid":        "device-4",
						"name":       "Device 4",
						"tenant_id":  "tenant-1",
						"status":     "accepted",
						"created_at": timestamp,
					},
					{
						"uid":        "device-5",
						"name":       "Device 5",
						"tenant_id":  "tenant-1",
						"status":     "pending",
						"created_at": timestamp,
					},
					// tenant-2: 1 removed device
					{
						"uid":               "device-6",
						"name":              "Device 6",
						"tenant_id":         "tenant-2",
						"status":            "removed",
						"status_updated_at": timestamp,
						"created_at":        timestamp,
					},
					// tenant-2: non-removed devices
					{
						"uid":        "device-7",
						"name":       "Device 7",
						"tenant_id":  "tenant-2",
						"status":     "accepted",
						"created_at": timestamp,
					},
					// tenant-3: no removed devices, only accepted
					{
						"uid":        "device-8",
						"name":       "Device 8",
						"tenant_id":  "tenant-3",
						"status":     "accepted",
						"created_at": timestamp,
					},
				}

				if _, err := c.Database("test").Collection("namespaces").InsertMany(ctx, []any{namespaces[0], namespaces[1], namespaces[2]}); err != nil {
					return err
				}

				rawDevices := []any{devices[0], devices[1], devices[2], devices[3], devices[4], devices[5], devices[6], devices[7]}
				_, err := c.Database("test").Collection("devices").InsertMany(ctx, rawDevices)

				return err
			},
			verify: func(tt *testing.T) {
				namespace1 := make(map[string]any)
				require.NoError(tt, c.Database("test").Collection("namespaces").FindOne(ctx, bson.M{"tenant_id": "tenant-1"}).Decode(&namespace1))
				assert.Equal(tt, int32(5), namespace1["devices_accepted_count"])
				assert.Equal(tt, int32(2), namespace1["devices_pending_count"])
				assert.Equal(tt, int32(1), namespace1["devices_rejected_count"])
				assert.Equal(tt, int32(3), namespace1["devices_removed_count"])

				namespace2 := make(map[string]any)
				require.NoError(tt, c.Database("test").Collection("namespaces").FindOne(ctx, bson.M{"tenant_id": "tenant-2"}).Decode(&namespace2))
				assert.Equal(tt, int32(3), namespace2["devices_accepted_count"])
				assert.Equal(tt, int32(0), namespace2["devices_pending_count"])
				assert.Equal(tt, int32(0), namespace2["devices_rejected_count"])
				assert.Equal(tt, int32(1), namespace2["devices_removed_count"])

				namespace3 := make(map[string]any)
				require.NoError(tt, c.Database("test").Collection("namespaces").FindOne(ctx, bson.M{"tenant_id": "tenant-3"}).Decode(&namespace3))
				assert.Equal(tt, int32(0), namespace3["devices_accepted_count"])
				assert.Equal(tt, int32(0), namespace3["devices_pending_count"])
				assert.Equal(tt, int32(0), namespace3["devices_rejected_count"])
				assert.Equal(tt, int32(0), namespace3["devices_removed_count"])
			},
		},
		{
			description: "succeeds initializing devices_removed_count to 0 when no removed devices exist",
			setup: func() error {
				timestamp := time.Now()

				namespaces := []bson.M{
					{
						"tenant_id":              "tenant-1",
						"devices_accepted_count": 2,
						"devices_pending_count":  1,
						"devices_rejected_count": 0,
					},
				}

				// Create devices with no "removed" status
				devices := []bson.M{
					{
						"uid":        "device-1",
						"name":       "Device 1",
						"tenant_id":  "tenant-1",
						"status":     "accepted",
						"created_at": timestamp,
					},
					{
						"uid":        "device-2",
						"name":       "Device 2",
						"tenant_id":  "tenant-1",
						"status":     "pending",
						"created_at": timestamp,
					},
				}

				if _, err := c.Database("test").Collection("namespaces").InsertMany(ctx, []any{namespaces[0]}); err != nil {
					return err
				}

				_, err := c.Database("test").Collection("devices").InsertMany(ctx, []any{devices[0], devices[1]})

				return err
			},
			verify: func(tt *testing.T) {
				namespace1 := make(map[string]any)
				require.NoError(tt, c.Database("test").Collection("namespaces").FindOne(ctx, bson.M{"tenant_id": "tenant-1"}).Decode(&namespace1))
				assert.Equal(tt, int32(2), namespace1["devices_accepted_count"])
				assert.Equal(tt, int32(1), namespace1["devices_pending_count"])
				assert.Equal(tt, int32(0), namespace1["devices_rejected_count"])
				assert.Equal(tt, int32(0), namespace1["devices_removed_count"])
			},
		},
		{
			description: "succeeds with empty collections",
			setup: func() error {
				namespaces := []bson.M{
					{
						"tenant_id":              "tenant-1",
						"devices_accepted_count": 0,
						"devices_pending_count":  0,
						"devices_rejected_count": 0,
					},
				}

				_, err := c.Database("test").Collection("namespaces").InsertMany(ctx, []any{namespaces[0]})

				return err
			},
			verify: func(tt *testing.T) {
				namespace1 := make(map[string]any)
				require.NoError(tt, c.Database("test").Collection("namespaces").FindOne(ctx, bson.M{"tenant_id": "tenant-1"}).Decode(&namespace1))
				assert.Equal(tt, int32(0), namespace1["devices_accepted_count"])
				assert.Equal(tt, int32(0), namespace1["devices_pending_count"])
				assert.Equal(tt, int32(0), namespace1["devices_rejected_count"])
				assert.Equal(tt, int32(0), namespace1["devices_removed_count"])
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			tt.Cleanup(func() { assert.NoError(tt, srv.Reset()) })

			require.NoError(tt, tc.setup())
			migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[103])
			require.NoError(tt, migrates.Up(ctx, migrate.AllAvailable))
			tc.verify(tt)
		})
	}
}

func TestMigration104Down(t *testing.T) {
	ctx := context.Background()

	cases := []struct {
		description string
		setup       func() error
		verify      func(tt *testing.T)
	}{
		{
			description: "succeeds removing devices_removed_count field",
			setup: func() error {
				namespaces := []bson.M{
					{
						"tenant_id":              "tenant-1",
						"devices_accepted_count": 42,
						"devices_pending_count":  5,
						"devices_rejected_count": 3,
						"devices_removed_count":  7,
					},
					{
						"tenant_id":              "tenant-2",
						"devices_accepted_count": 10,
						"devices_removed_count":  2,
					},
				}
				_, err := c.Database("test").Collection("namespaces").InsertMany(ctx, []any{namespaces[0], namespaces[1]})

				return err
			},
			verify: func(tt *testing.T) {
				namespace1 := make(map[string]any)
				require.NoError(tt, c.Database("test").Collection("namespaces").FindOne(ctx, bson.M{"tenant_id": "tenant-1"}).Decode(&namespace1))
				_, hasRemovedCount := namespace1["devices_removed_count"]
				assert.False(tt, hasRemovedCount)
				assert.Equal(tt, int32(42), namespace1["devices_accepted_count"])
				assert.Equal(tt, int32(5), namespace1["devices_pending_count"])
				assert.Equal(tt, int32(3), namespace1["devices_rejected_count"])

				namespace2 := make(map[string]any)
				require.NoError(tt, c.Database("test").Collection("namespaces").FindOne(ctx, bson.M{"tenant_id": "tenant-2"}).Decode(&namespace2))
				_, hasRemovedCount2 := namespace2["devices_removed_count"]
				assert.False(tt, hasRemovedCount2)
				assert.Equal(tt, int32(10), namespace2["devices_accepted_count"])
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			tt.Cleanup(func() { assert.NoError(tt, srv.Reset()) })

			require.NoError(tt, tc.setup())
			migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[103])
			require.NoError(tt, migrates.Up(ctx, migrate.AllAvailable))
			require.NoError(tt, migrates.Down(ctx, migrate.AllAvailable))
			tc.verify(tt)
		})
	}
}
