package migrations

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
)

func TestMigration120Up(t *testing.T) {
	ctx := context.Background()

	cases := []struct {
		description string
		setup       func() error
		verify      func(tt *testing.T)
	}{
		{
			description: "succeeds syncing counters and converting types",
			setup: func() error {
				namespaces := []bson.M{
					{
						"tenant_id":              "tenant-1",
						"name":                   "namespace-1",
						"devices_accepted_count": int32(99),
						"devices_pending_count":  int32(99),
						"devices_rejected_count": int32(99),
						"devices_removed_count":  int32(99),
					},
					{
						"tenant_id":              "tenant-2",
						"name":                   "namespace-2",
						"devices_accepted_count": int32(0),
						"devices_pending_count":  int32(0),
						"devices_rejected_count": int32(0),
						"devices_removed_count":  int32(0),
					},
				}

				if _, err := c.Database("test").Collection("namespaces").InsertMany(ctx, []any{namespaces[0], namespaces[1]}); err != nil {
					return err
				}

				devices := []bson.M{
					{"tenant_id": "tenant-1", "uid": "dev-1", "status": "accepted"},
					{"tenant_id": "tenant-1", "uid": "dev-2", "status": "accepted"},
					{"tenant_id": "tenant-1", "uid": "dev-3", "status": "pending"},
					{"tenant_id": "tenant-1", "uid": "dev-4", "status": "rejected"},
					{"tenant_id": "tenant-1", "uid": "dev-5", "status": "removed"},
					{"tenant_id": "tenant-2", "uid": "dev-6", "status": "accepted"},
				}

				_, err := c.Database("test").Collection("devices").InsertMany(ctx, []any{
					devices[0], devices[1], devices[2], devices[3], devices[4], devices[5],
				})

				return err
			},
			verify: func(tt *testing.T) {
				ns1 := make(map[string]any)
				require.NoError(tt, c.Database("test").Collection("namespaces").FindOne(ctx, bson.M{"tenant_id": "tenant-1"}).Decode(&ns1))
				assert.Equal(tt, int64(2), ns1["devices_accepted_count"])
				assert.Equal(tt, int64(1), ns1["devices_pending_count"])
				assert.Equal(tt, int64(1), ns1["devices_rejected_count"])
				assert.Equal(tt, int64(1), ns1["devices_removed_count"])

				ns2 := make(map[string]any)
				require.NoError(tt, c.Database("test").Collection("namespaces").FindOne(ctx, bson.M{"tenant_id": "tenant-2"}).Decode(&ns2))
				assert.Equal(tt, int64(1), ns2["devices_accepted_count"])
				assert.Equal(tt, int64(0), ns2["devices_pending_count"])
				assert.Equal(tt, int64(0), ns2["devices_rejected_count"])
				assert.Equal(tt, int64(0), ns2["devices_removed_count"])
			},
		},
		{
			description: "succeeds with namespaces that have no devices",
			setup: func() error {
				namespace := bson.M{
					"tenant_id":              "tenant-empty",
					"name":                   "empty-namespace",
					"devices_accepted_count": int32(10),
					"devices_pending_count":  int32(5),
					"devices_rejected_count": int32(3),
					"devices_removed_count":  int32(1),
				}
				_, err := c.Database("test").Collection("namespaces").InsertOne(ctx, namespace)

				return err
			},
			verify: func(tt *testing.T) {
				ns := make(map[string]any)
				require.NoError(tt, c.Database("test").Collection("namespaces").FindOne(ctx, bson.M{"tenant_id": "tenant-empty"}).Decode(&ns))
				assert.Equal(tt, int64(0), ns["devices_accepted_count"])
				assert.Equal(tt, int64(0), ns["devices_pending_count"])
				assert.Equal(tt, int64(0), ns["devices_rejected_count"])
				assert.Equal(tt, int64(0), ns["devices_removed_count"])
			},
		},
		{
			description: "succeeds with removed devices counted correctly",
			setup: func() error {
				namespace := bson.M{
					"tenant_id":              "tenant-removed",
					"name":                   "removed-namespace",
					"devices_accepted_count": int32(0),
					"devices_pending_count":  int32(0),
					"devices_rejected_count": int32(0),
					"devices_removed_count":  int32(0),
				}
				if _, err := c.Database("test").Collection("namespaces").InsertOne(ctx, namespace); err != nil {
					return err
				}

				devices := []bson.M{
					{"tenant_id": "tenant-removed", "uid": "dev-r1", "status": "removed"},
					{"tenant_id": "tenant-removed", "uid": "dev-r2", "status": "removed"},
					{"tenant_id": "tenant-removed", "uid": "dev-r3", "status": "removed"},
					{"tenant_id": "tenant-removed", "uid": "dev-a1", "status": "accepted"},
				}

				_, err := c.Database("test").Collection("devices").InsertMany(ctx, []any{
					devices[0], devices[1], devices[2], devices[3],
				})

				return err
			},
			verify: func(tt *testing.T) {
				ns := make(map[string]any)
				require.NoError(tt, c.Database("test").Collection("namespaces").FindOne(ctx, bson.M{"tenant_id": "tenant-removed"}).Decode(&ns))
				assert.Equal(tt, int64(1), ns["devices_accepted_count"])
				assert.Equal(tt, int64(0), ns["devices_pending_count"])
				assert.Equal(tt, int64(0), ns["devices_rejected_count"])
				assert.Equal(tt, int64(3), ns["devices_removed_count"])
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			tt.Cleanup(func() { assert.NoError(tt, srv.Reset()) })
			require.NoError(tt, tc.setup())

			migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[119])
			require.NoError(tt, migrates.Up(ctx, migrate.AllAvailable))
			tc.verify(tt)
		})
	}
}
