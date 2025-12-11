package migrations

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestMigration118Up(t *testing.T) {
	ctx := context.Background()

	cases := []struct {
		description string
		setup       func() error
		verify      func(tt *testing.T)
	}{
		{
			description: "succeeds migrating namespace members to membership_invitations collection",
			setup: func() error {
				ownerID := primitive.NewObjectID()
				memberID1 := primitive.NewObjectID()
				memberID2 := primitive.NewObjectID()

				namespaces := []bson.M{
					{
						"_id":       primitive.NewObjectID(),
						"name":      "test-namespace-1",
						"owner":     ownerID,
						"tenant_id": "tenant-1",
						"members": bson.A{
							bson.M{
								"id":         ownerID,
								"added_at":   primitive.NewDateTimeFromTime(primitive.NewObjectID().Timestamp()),
								"role":       "owner",
								"status":     "accepted",
								"expires_at": nil,
							},
							bson.M{
								"id":         memberID1,
								"added_at":   primitive.NewDateTimeFromTime(primitive.NewObjectID().Timestamp()),
								"role":       "observer",
								"status":     "pending",
								"expires_at": primitive.NewDateTimeFromTime(primitive.NewObjectID().Timestamp().Add(7 * 24 * 60 * 60 * 1000)),
							},
							bson.M{
								"id":         memberID2,
								"added_at":   primitive.NewDateTimeFromTime(primitive.NewObjectID().Timestamp()),
								"role":       "administrator",
								"status":     "accepted",
								"expires_at": nil,
							},
						},
					},
				}

				_, err := c.Database("test").Collection("namespaces").InsertMany(ctx, []any{namespaces[0]})

				return err
			},
			verify: func(tt *testing.T) {
				cursor, err := c.Database("test").Collection("membership_invitations").Find(ctx, bson.M{})
				require.NoError(tt, err)

				invitations := make([]bson.M, 0)
				require.NoError(tt, cursor.All(ctx, &invitations))
				require.Equal(tt, 2, len(invitations))

				ownerFound := false
				for _, invitation := range invitations {
					require.NotNil(tt, invitation["_id"])
					require.Equal(tt, "tenant-1", invitation["tenant_id"])
					require.NotNil(tt, invitation["user_id"])
					require.NotNil(tt, invitation["invited_by"])
					require.NotNil(tt, invitation["role"])
					require.NotNil(tt, invitation["status"])
					require.NotNil(tt, invitation["created_at"])
					require.NotNil(tt, invitation["updated_at"])
					require.NotNil(tt, invitation["status_updated_at"])
					require.Equal(tt, int32(1), invitation["invitations"])

					require.NotEqual(tt, "owner", invitation["role"])
					if invitation["role"] == "owner" {
						ownerFound = true
					}
				}
				require.False(tt, ownerFound, "Owner should not have an invitation created")

				namespaceCursor, err := c.Database("test").Collection("namespaces").Find(ctx, bson.M{"tenant_id": "tenant-1"})
				require.NoError(tt, err)

				namespaces := make([]bson.M, 0)
				require.NoError(tt, namespaceCursor.All(ctx, &namespaces))
				require.Equal(tt, 1, len(namespaces))

				namespace := namespaces[0]
				members, ok := namespace["members"].(bson.A)
				require.True(tt, ok)
				require.Equal(tt, 2, len(members))

				for _, m := range members {
					member, ok := m.(bson.M)
					require.True(tt, ok)
					require.NotNil(tt, member["id"])
					require.NotNil(tt, member["added_at"])
					require.NotNil(tt, member["role"])
					require.Nil(tt, member["status"])
					require.Nil(tt, member["expires_at"])
				}
			},
		},
		{
			description: "handles namespace with no members gracefully",
			setup: func() error {
				namespaces := []bson.M{
					{
						"_id":       primitive.NewObjectID(),
						"name":      "empty-namespace",
						"owner":     primitive.NewObjectID(),
						"tenant_id": "tenant-empty",
						"members":   bson.A{},
					},
				}

				_, err := c.Database("test").Collection("namespaces").InsertMany(ctx, []any{namespaces[0]})

				return err
			},
			verify: func(tt *testing.T) {
				count, err := c.Database("test").Collection("membership_invitations").CountDocuments(ctx, bson.M{})
				require.NoError(tt, err)
				require.Equal(tt, int64(0), count)

				namespaceCount, err := c.Database("test").Collection("namespaces").CountDocuments(ctx, bson.M{"tenant_id": "tenant-empty"})
				require.NoError(tt, err)
				require.Equal(tt, int64(1), namespaceCount)
			},
		},
		{
			description: "handles empty namespaces collection gracefully",
			setup: func() error {
				return nil
			},
			verify: func(tt *testing.T) {
				count, err := c.Database("test").Collection("membership_invitations").CountDocuments(ctx, bson.M{})
				require.NoError(tt, err)
				require.Equal(tt, int64(0), count)

				namespaceCount, err := c.Database("test").Collection("namespaces").CountDocuments(ctx, bson.M{})
				require.NoError(tt, err)
				require.Equal(tt, int64(0), namespaceCount)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			tt.Cleanup(func() { require.NoError(tt, srv.Reset()) })

			require.NoError(tt, tc.setup())
			migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[117])
			require.NoError(tt, migrates.Up(ctx, migrate.AllAvailable))
			tc.verify(tt)
		})
	}
}
