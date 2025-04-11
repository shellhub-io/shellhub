package migrations

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
)

func TestMigration98Up(t *testing.T) {
	ctx := context.Background()

	cases := []struct {
		description string
		setup       func() error
		verify      func(tt *testing.T)
	}{
		{
			description: "Session with empty type should be updated to personal",
			setup: func() error {
				_, err := c.
					Database("test").
					Collection("namespaces").
					InsertOne(ctx, bson.M{
						"tenant_id": "empty-type-test",
						"type":      "",
					})

				return err
			},
			verify: func(tt *testing.T) {
				query := c.
					Database("test").
					Collection("namespaces").
					FindOne(ctx, bson.M{"tenant_id": "empty-type-test"})
				namespace := make(map[string]any)
				require.NoError(tt, query.Decode(&namespace))

				assert.Equal(tt, "personal", namespace["type"],
					"Session type should be updated to 'personal' when originally empty")
			},
		},
		{
			description: "Session with non-empty type should remain unchanged",
			setup: func() error {
				_, err := c.
					Database("test").
					Collection("namespaces").
					InsertOne(ctx, bson.M{
						"tenant_id": "existing-type-test",
						"type":      "existing",
					})

				return err
			},
			verify: func(tt *testing.T) {
				query := c.
					Database("test").
					Collection("namespaces").
					FindOne(ctx, bson.M{"tenant_id": "existing-type-test"})
				namespace := make(map[string]any)
				require.NoError(tt, query.Decode(&namespace))

				assert.Equal(tt, "existing", namespace["type"],
					"Session type should remain unchanged when not empty")
			},
		},
		{
			description: "Multiple namespaces with empty type should be updated",
			setup: func() error {
				_, err := c.
					Database("test").
					Collection("namespaces").
					InsertMany(ctx, []any{
						bson.M{
							"tenant_id": "multi-empty-1",
							"type":      "",
						},
						bson.M{
							"tenant_id": "multi-empty-2",
							"type":      "",
						},
						bson.M{
							"tenant_id": "multi-existing",
							"type":      "existing",
						},
					})

				return err
			},
			verify: func(tt *testing.T) {
				query := c.
					Database("test").
					Collection("namespaces").
					FindOne(ctx, bson.M{"tenant_id": "multi-empty-1"})
				namespace1 := make(map[string]any)
				require.NoError(tt, query.Decode(&namespace1))
				assert.Equal(tt, "personal", namespace1["type"])

				query = c.
					Database("test").
					Collection("namespaces").
					FindOne(ctx, bson.M{"tenant_id": "multi-empty-2"})
				namespace2 := make(map[string]any)
				require.NoError(tt, query.Decode(&namespace2))
				assert.Equal(tt, "personal", namespace2["type"])

				query = c.
					Database("test").
					Collection("namespaces").
					FindOne(ctx, bson.M{"tenant_id": "multi-existing"})
				namespaceExisting := make(map[string]any)
				require.NoError(tt, query.Decode(&namespaceExisting))
				assert.Equal(tt, "existing", namespaceExisting["type"])
			},
		},
		{
			description: "No namespaces with empty type should handle gracefully",
			setup: func() error {
				return nil
			},
			verify: func(tt *testing.T) {
				count, err := c.
					Database("test").
					Collection("namespaces").
					CountDocuments(ctx, bson.M{"type": ""})
				require.NoError(tt, err)
				assert.Equal(tt, int64(0), count)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			tt.Cleanup(func() {
				assert.NoError(tt, srv.Reset())
			})

			require.NoError(tt, tc.setup())

			migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[97])
			require.NoError(tt, migrates.Up(ctx, migrate.AllAvailable))

			tc.verify(tt)
		})
	}
}
