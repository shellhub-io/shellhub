package migrations

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestMigration115Up(t *testing.T) {
	ctx := context.Background()

	cases := []struct {
		description string
		setup       func() error
		verify      func(tt *testing.T)
	}{
		{
			description: "succeeds migrating invited users to user_invitations collection",
			setup: func() error {
				users := []bson.M{
					{
						"_id":        primitive.NewObjectID(),
						"email":      "invited1@test.com",
						"created_at": primitive.NewDateTimeFromTime(primitive.NewObjectID().Timestamp()),
						"status":     "invited",
						"name":       nil,
						"username":   nil,
					},
					{
						"_id":        primitive.NewObjectID(),
						"email":      "invited2@test.com",
						"created_at": primitive.NewDateTimeFromTime(primitive.NewObjectID().Timestamp()),
						"status":     "invited",
						"name":       nil,
						"username":   nil,
					},
					{
						"_id":        primitive.NewObjectID(),
						"email":      "confirmed@test.com",
						"created_at": primitive.NewDateTimeFromTime(primitive.NewObjectID().Timestamp()),
						"status":     "confirmed",
						"name":       "John Doe",
						"username":   "johndoe",
					},
				}

				_, err := c.Database("test").Collection("users").InsertMany(ctx, []any{users[0], users[1], users[2]})

				return err
			},
			verify: func(tt *testing.T) {
				cursor, err := c.Database("test").Collection("user_invitations").Find(ctx, bson.M{})
				require.NoError(tt, err)

				invitations := make([]bson.M, 0)
				require.NoError(tt, cursor.All(ctx, &invitations))
				require.Equal(tt, 2, len(invitations))

				for _, invitation := range invitations {
					require.NotNil(tt, invitation["_id"])
					require.NotNil(tt, invitation["email"])
					require.NotNil(tt, invitation["created_at"])
					require.NotNil(tt, invitation["updated_at"])
					require.Contains(tt, []string{"invited1@test.com", "invited2@test.com"}, invitation["email"])
					require.Equal(tt, int32(1), invitation["invitations"])
					require.Equal(tt, "pending", invitation["status"])

					require.Nil(tt, invitation["name"])
					require.Nil(tt, invitation["username"])
				}

				userCursor, err := c.Database("test").Collection("users").Find(ctx, bson.M{"status": "invited"})
				require.NoError(tt, err)

				invitedUsers := make([]bson.M, 0)
				require.NoError(tt, userCursor.All(ctx, &invitedUsers))
				require.Equal(tt, 0, len(invitedUsers))

				confirmedCursor, err := c.Database("test").Collection("users").Find(ctx, bson.M{"status": "confirmed"})
				require.NoError(tt, err)

				confirmedUsers := make([]bson.M, 0)
				require.NoError(tt, confirmedCursor.All(ctx, &confirmedUsers))
				require.Equal(tt, 1, len(confirmedUsers))
				require.Equal(tt, "confirmed@test.com", confirmedUsers[0]["email"])
			},
		},
		{
			description: "handles empty users collection gracefully",
			setup: func() error {
				return nil
			},
			verify: func(tt *testing.T) {
				count, err := c.Database("test").Collection("user_invitations").CountDocuments(ctx, bson.M{})
				require.NoError(tt, err)
				require.Equal(tt, int64(0), count)

				userCount, err := c.Database("test").Collection("users").CountDocuments(ctx, bson.M{})
				require.NoError(tt, err)
				require.Equal(tt, int64(0), userCount)
			},
		},
		{
			description: "handles users collection with no invited users",
			setup: func() error {
				users := []bson.M{
					{
						"_id":        primitive.NewObjectID(),
						"email":      "confirmed1@test.com",
						"created_at": primitive.NewDateTimeFromTime(primitive.NewObjectID().Timestamp()),
						"status":     "confirmed",
						"name":       "Jane Doe",
						"username":   "janedoe",
					},
					{
						"_id":        primitive.NewObjectID(),
						"email":      "not-confirmed@test.com",
						"created_at": primitive.NewDateTimeFromTime(primitive.NewObjectID().Timestamp()),
						"status":     "not-confirmed",
						"name":       "Bob Smith",
						"username":   "bobsmith",
					},
				}

				_, err := c.Database("test").Collection("users").InsertMany(ctx, []any{users[0], users[1]})

				return err
			},
			verify: func(tt *testing.T) {
				invitationCount, err := c.Database("test").Collection("user_invitations").CountDocuments(ctx, bson.M{})
				require.NoError(tt, err)
				require.Equal(tt, int64(0), invitationCount)

				userCount, err := c.Database("test").Collection("users").CountDocuments(ctx, bson.M{})
				require.NoError(tt, err)
				require.Equal(tt, int64(2), userCount)

				invitedCount, err := c.Database("test").Collection("users").CountDocuments(ctx, bson.M{"status": "invited"})
				require.NoError(tt, err)
				require.Equal(tt, int64(0), invitedCount)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			tt.Cleanup(func() { require.NoError(tt, srv.Reset()) })

			require.NoError(tt, tc.setup())
			migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[114])
			require.NoError(tt, migrates.Up(ctx, migrate.AllAvailable))
			tc.verify(tt)
		})
	}
}

func TestMigration115Down(t *testing.T) {
	ctx := context.Background()

	cases := []struct {
		description string
		setup       func() error
		verify      func(tt *testing.T)
	}{
		{
			description: "succeeds reverting user_invitations back to users collection with invited status",
			setup: func() error {
				invitations := []bson.M{
					{
						"_id":         primitive.NewObjectID(),
						"email":       "invited1@test.com",
						"created_at":  primitive.NewDateTimeFromTime(primitive.NewObjectID().Timestamp()),
						"updated_at":  primitive.NewDateTimeFromTime(primitive.NewObjectID().Timestamp()),
						"invitations": 1,
						"status":      "pending",
					},
					{
						"_id":         primitive.NewObjectID(),
						"email":       "invited2@test.com",
						"created_at":  primitive.NewDateTimeFromTime(primitive.NewObjectID().Timestamp()),
						"updated_at":  primitive.NewDateTimeFromTime(primitive.NewObjectID().Timestamp()),
						"invitations": 2,
						"status":      "accepted",
					},
				}

				_, err := c.Database("test").Collection("user_invitations").InsertMany(ctx, []any{invitations[0], invitations[1]})
				if err != nil {
					return err
				}

				_, err = c.Database("test").Collection("user_invitations").CountDocuments(ctx, bson.M{})
				if err != nil {
					return err
				}

				return nil
			},
			verify: func(tt *testing.T) {
				cursor, err := c.Database("test").Collection("users").Find(ctx, bson.M{"status": "invited"})
				require.NoError(tt, err)

				users := make([]bson.M, 0)
				require.NoError(tt, cursor.All(ctx, &users))
				require.Equal(tt, 2, len(users))

				for _, user := range users {
					require.NotNil(tt, user["_id"])
					require.NotNil(tt, user["email"])
					require.NotNil(tt, user["created_at"])
					require.Contains(tt, []string{"invited1@test.com", "invited2@test.com"}, user["email"])
					require.Equal(tt, "invited", user["status"])
					require.Nil(tt, user["name"])
					require.Nil(tt, user["username"])
					require.Nil(tt, user["last_login"])
				}

				count, err := c.Database("test").Collection("user_invitations").CountDocuments(ctx, bson.M{})
				require.NoError(tt, err)
				require.Equal(tt, int64(0), count)
			},
		},
		{
			description: "handles empty user_invitations collection gracefully",
			setup: func() error {
				return nil
			},
			verify: func(tt *testing.T) {
				userCount, err := c.Database("test").Collection("users").CountDocuments(ctx, bson.M{})
				require.NoError(tt, err)
				require.Equal(tt, int64(0), userCount)

				invitationCount, err := c.Database("test").Collection("user_invitations").CountDocuments(ctx, bson.M{})
				require.NoError(tt, err)
				require.Equal(tt, int64(0), invitationCount)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			tt.Cleanup(func() { require.NoError(tt, srv.Reset()) })

			require.NoError(tt, tc.setup())
			migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[114])
			require.NoError(tt, migrates.Up(ctx, migrate.AllAvailable))
			require.NoError(tt, migrates.Down(ctx, migrate.AllAvailable))
			tc.verify(tt)
		})
	}
}
