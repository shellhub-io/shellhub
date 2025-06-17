package migrations

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
)

func TestMigration102Up(t *testing.T) {
	ctx := context.Background()

	cases := []struct {
		description string
		setup       func() error
		verify      func(tt *testing.T)
	}{
		{
			description: "succeeds removing legacy devices_count field",
			setup: func() error {
				namespaces := []bson.M{
					{
						"tenant_id":              "tenant-1",
						"devices_count":          42,
						"devices_accepted_count": 42,
						"devices_pending_count":  0,
						"devices_rejected_count": 0,
					},
					{
						"tenant_id":              "tenant-2",
						"devices_accepted_count": 5,
					},
				}
				_, err := c.Database("test").Collection("namespaces").InsertMany(ctx, []any{namespaces[0], namespaces[1]})

				return err
			},
			verify: func(tt *testing.T) {
				namespace1 := make(map[string]any)
				require.NoError(tt, c.Database("test").Collection("namespaces").FindOne(ctx, bson.M{"tenant_id": "tenant-1"}).Decode(&namespace1))
				_, hasLegacyCount := namespace1["devices_count"]
				assert.False(tt, hasLegacyCount)
				assert.Equal(tt, int32(42), namespace1["devices_accepted_count"])
				assert.Equal(tt, int32(0), namespace1["devices_pending_count"])
				assert.Equal(tt, int32(0), namespace1["devices_rejected_count"])

				namespace2 := make(map[string]any)
				require.NoError(tt, c.Database("test").Collection("namespaces").FindOne(ctx, bson.M{"tenant_id": "tenant-2"}).Decode(&namespace2))
				_, hasLegacyCount3 := namespace2["devices_count"]
				assert.False(tt, hasLegacyCount3)
				assert.Equal(tt, int32(5), namespace2["devices_accepted_count"])
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
