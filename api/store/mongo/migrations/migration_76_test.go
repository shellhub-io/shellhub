package migrations

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/envs"
	envmock "github.com/shellhub-io/shellhub/pkg/envs/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestMigration76Up(t *testing.T) {
	ctx := context.Background()

	mock := &envmock.Backend{}
	envs.DefaultBackend = mock

	cases := []struct {
		setup       func(primitive.ObjectID) error
		description string
	}{
		{
			description: "Success to apply up on migration 76",
			setup: func(objID primitive.ObjectID) error {
				_, err := c.
					Database("test").
					Collection("users").
					InsertOne(ctx, map[string]interface{}{
						"_id":        objID,
						"namespaces": 1,
					})

				return err
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.description, func(tt *testing.T) {
			tt.Cleanup(func() {
				assert.NoError(tt, srv.Reset())
			})

			objID := primitive.NewObjectID()
			assert.NoError(tt, tc.setup(objID))

			migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[75])
			require.NoError(tt, migrates.Up(context.Background(), migrate.AllAvailable))

			query := c.
				Database("test").
				Collection("users").
				FindOne(context.TODO(), bson.M{"_id": objID})

			user := make(map[string]interface{})
			require.NoError(tt, query.Decode(&user))

			_, ok := user["namespaces"]
			require.Equal(tt, false, ok)
		})
	}
}

func TestMigration76Down(t *testing.T) {
	db := c.Database("test")
	ctx := context.Background()

	mock := &envmock.Backend{}
	envs.DefaultBackend = mock

	cases := []struct {
		setup       func(primitive.ObjectID) error
		description string
		expected    int32
	}{
		{
			description: "Success to apply up on migration 76",
			setup: func(objID primitive.ObjectID) error {
				_, err := db.Collection("users").InsertOne(ctx, map[string]interface{}{"_id": objID})
				if err != nil {
					return err
				}

				namespaces := []map[string]interface{}{
					{
						"_id": primitive.NewObjectID(),
						"members": []bson.M{
							{"id": objID.Hex(), "role": "owner"},
							{"id": "000000000000000000000000", "role": "observer"},
						},
					},
					{
						"_id": primitive.NewObjectID(),
						"members": []bson.M{
							{"id": objID.Hex(), "role": "owner"},
							{"id": "000000000000000000000000", "role": "observer"},
						},
					},
					{
						"_id": primitive.NewObjectID(),
						"members": []bson.M{
							{"id": objID.Hex(), "role": "observer"},
							{"id": "000000000000000000000000", "role": "owner"},
						},
					},
				}

				for _, n := range namespaces {
					_, err = db.Collection("namespaces").InsertOne(ctx, n)
					if err != nil {
						return err
					}
				}

				return nil
			},
			expected: 2,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.description, func(tt *testing.T) {
			tt.Cleanup(func() {
				assert.NoError(tt, srv.Reset())
			})

			objID := primitive.NewObjectID()
			assert.NoError(tt, tc.setup(objID))

			migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[75])
			require.NoError(tt, migrates.Up(context.Background(), migrate.AllAvailable))
			require.NoError(tt, migrates.Down(context.Background(), migrate.AllAvailable))

			query := c.
				Database("test").
				Collection("users").
				FindOne(context.TODO(), bson.M{"_id": objID})

			user := make(map[string]interface{})
			require.NoError(tt, query.Decode(&user))

			count, ok := user["namespaces"]
			require.Equal(tt, true, ok)
			require.Equal(tt, tc.expected, count.(int32))
		})
	}
}
