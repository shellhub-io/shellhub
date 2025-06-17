package migrations

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
)

func TestMigration101Up(t *testing.T) {
	ctx := context.Background()

	cases := []struct {
		description string
		setup       func() error
		verify      func(tt *testing.T)
	}{
		{
			description: "succeeds with devices and namespaces",
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

				devices := []bson.M{
					// tenant-1: 3 accepted, 2 pending, 1 rejected
					{"tenant_id": "tenant-1", "uid": "dev-1", "status": "accepted"},
					{"tenant_id": "tenant-1", "uid": "dev-2", "status": "accepted"},
					{"tenant_id": "tenant-1", "uid": "dev-3", "status": "accepted"},
					{"tenant_id": "tenant-1", "uid": "dev-4", "status": "pending"},
					{"tenant_id": "tenant-1", "uid": "dev-5", "status": "pending"},
					{"tenant_id": "tenant-1", "uid": "dev-6", "status": "rejected"},

					// tenant-2: 1 accepted, 0 pending, 2 rejected
					{"tenant_id": "tenant-2", "uid": "dev-7", "status": "accepted"},
					{"tenant_id": "tenant-2", "uid": "dev-8", "status": "rejected"},
					{"tenant_id": "tenant-2", "uid": "dev-9", "status": "rejected"},

					// tenant-3: no devices (test empty case)
				}

				_, err = c.
					Database("test").
					Collection("devices").
					InsertMany(ctx, []any{devices[0], devices[1], devices[2], devices[3], devices[4], devices[5], devices[6], devices[7], devices[8]})

				return err
			},
			verify: func(tt *testing.T) {
				namespace1 := make(map[string]any)
				require.NoError(tt, c.Database("test").Collection("namespaces").FindOne(ctx, bson.M{"tenant_id": "tenant-1"}).Decode(&namespace1))
				assert.Equal(tt, int32(3), namespace1["devices_accepted_count"].(int32))
				assert.Equal(tt, int32(2), namespace1["devices_pending_count"].(int32))
				assert.Equal(tt, int32(1), namespace1["devices_rejected_count"].(int32))

				namespace2 := make(map[string]any)
				require.NoError(tt, c.Database("test").Collection("namespaces").FindOne(ctx, bson.M{"tenant_id": "tenant-2"}).Decode(&namespace2))
				assert.Equal(tt, int32(1), namespace2["devices_accepted_count"].(int32))
				assert.Equal(tt, int32(0), namespace2["devices_pending_count"].(int32))
				assert.Equal(tt, int32(2), namespace2["devices_rejected_count"].(int32))

				namespace3 := make(map[string]any)
				require.NoError(tt, c.Database("test").Collection("namespaces").FindOne(ctx, bson.M{"tenant_id": "tenant-3"}).Decode(&namespace3))
				assert.Equal(tt, int32(0), namespace3["devices_accepted_count"])
				assert.Equal(tt, int32(0), namespace3["devices_pending_count"])
				assert.Equal(tt, int32(0), namespace3["devices_rejected_count"])
			},
		},
		{
			description: "succeeds with namespaces but no devices",
			setup: func() error {
				namespace := bson.M{"tenant_id": "tenant-empty", "name": "empty-namespace"}
				_, err := c.Database("test").Collection("namespaces").InsertOne(ctx, namespace)

				return err
			},
			verify: func(tt *testing.T) {
				namespace := make(map[string]any)
				require.NoError(tt, c.Database("test").Collection("namespaces").FindOne(ctx, bson.M{"tenant_id": "tenant-empty"}).Decode(&namespace))
				assert.Equal(tt, int32(0), namespace["devices_accepted_count"])
				assert.Equal(tt, int32(0), namespace["devices_pending_count"])
				assert.Equal(tt, int32(0), namespace["devices_rejected_count"])
			},
		},
		{
			description: "succeeds with single status devices",
			setup: func() error {
				namespace := bson.M{
					"tenant_id": "tenant-single",
					"name":      "single-status-namespace",
				}
				_, err := c.Database("test").Collection("namespaces").InsertOne(ctx, namespace)
				if err != nil {
					return err
				}

				devices := []bson.M{
					{"tenant_id": "tenant-single", "uid": "dev-a", "status": "pending"},
					{"tenant_id": "tenant-single", "uid": "dev-b", "status": "pending"},
					{"tenant_id": "tenant-single", "uid": "dev-c", "status": "pending"},
				}

				_, err = c.Database("test").Collection("devices").InsertMany(ctx, []any{devices[0], devices[1], devices[2]})

				return err
			},
			verify: func(tt *testing.T) {
				namespace := make(map[string]any)
				require.NoError(tt, c.Database("test").Collection("namespaces").FindOne(ctx, bson.M{"tenant_id": "tenant-single"}).Decode(&namespace))
				assert.Equal(tt, int32(0), namespace["devices_accepted_count"])
				assert.Equal(tt, int32(3), namespace["devices_pending_count"])
				assert.Equal(tt, int32(0), namespace["devices_rejected_count"])
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			tt.Cleanup(func() { assert.NoError(tt, srv.Reset()) })
			require.NoError(tt, tc.setup())

			migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[100])
			require.NoError(tt, migrates.Up(ctx, migrate.AllAvailable))
			tc.verify(tt)
		})
	}
}

func TestMigration101Down(t *testing.T) {
	ctx := context.Background()

	cases := []struct {
		description string
		setup       func() error
		verify      func(tt *testing.T)
	}{
		{
			description: "succeeds removing device count fields",
			setup: func() error {
				_, err := c.
					Database("test").
					Collection("namespaces").
					InsertOne(
						ctx,
						bson.M{
							"tenant_id":              "tenant-down",
							"name":                   "test-namespace",
							"devices_accepted_count": 15,
							"devices_pending_count":  3,
							"devices_rejected_count": 2,
							"other_field":            "should_remain",
						},
					)

				return err
			},
			verify: func(tt *testing.T) {
				namespace := make(map[string]any)
				require.NoError(tt, c.Database("test").Collection("namespaces").FindOne(ctx, bson.M{"tenant_id": "tenant-down"}).Decode(&namespace))

				_, hasAccepted := namespace["devices_accepted_count"]
				_, hasPending := namespace["devices_pending_count"]
				_, hasRejected := namespace["devices_rejected_count"]

				assert.False(tt, hasAccepted)
				assert.False(tt, hasPending)
				assert.False(tt, hasRejected)
				assert.Equal(tt, "test-namespace", namespace["name"])
				assert.Equal(tt, "should_remain", namespace["other_field"])
			},
		},
		{
			description: "succeeds with multiple namespaces",
			setup: func() error {
				namespaces := []bson.M{
					{
						"tenant_id":              "tenant-1",
						"name":                   "namespace-1",
						"devices_accepted_count": 10,
						"devices_pending_count":  5,
					},
					{
						"tenant_id":              "tenant-2",
						"name":                   "namespace-2",
						"devices_rejected_count": 2,
					},
				}
				_, err := c.Database("test").Collection("namespaces").InsertMany(ctx, []any{namespaces[0], namespaces[1]})

				return err
			},
			verify: func(tt *testing.T) {
				namespace1 := make(map[string]any)
				require.NoError(tt, c.Database("test").Collection("namespaces").FindOne(ctx, bson.M{"tenant_id": "tenant-1"}).Decode(&namespace1))

				_, hasFields1 := namespace1["devices_accepted_count"]
				assert.False(tt, hasFields1)
				assert.Equal(tt, "namespace-1", namespace1["name"])

				namespace2 := make(map[string]any)
				require.NoError(tt, c.Database("test").Collection("namespaces").FindOne(ctx, bson.M{"tenant_id": "tenant-2"}).Decode(&namespace2))

				_, hasFields2 := namespace2["devices_rejected_count"]
				assert.False(tt, hasFields2)
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

			migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[100]) // index 100 for migration 101
			require.NoError(tt, migrates.Up(ctx, migrate.AllAvailable))
			require.NoError(tt, migrates.Down(ctx, migrate.AllAvailable))

			tc.verify(tt)
		})
	}
}
