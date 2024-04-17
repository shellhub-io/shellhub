package migrations

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/api/pkg/fixtures"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestMigration49(t *testing.T) {
	cases := []struct {
		description string
		Test        func(t *testing.T)
	}{
		{
			"Success to apply up on migration 49",
			func(t *testing.T) {
				t.Helper()

				user1ID, err := primitive.ObjectIDFromHex("507f1f77bcf86cd799439011")
				assert.NoError(t, err)
				user1 := &models.User{
					ID: user1ID.String(),
				}

				user2ID, err := primitive.ObjectIDFromHex("507f1f77bcf86cd799439012")
				assert.NoError(t, err)
				user2 := &models.User{
					ID: user2ID.String(),
				}

				namespace1 := &models.Namespace{
					Name:  "namespace1",
					Owner: user1ID.String(),
				}
				namespace2 := &models.Namespace{
					Name:  "namespace2",
					Owner: user1ID.String(),
				}
				namespace3 := &models.Namespace{
					Name:  "namespace3",
					Owner: user2ID.String(),
				}

				_, err = srv.Client().Database("test").Collection("users").InsertOne(context.TODO(), user1)
				assert.NoError(t, err)
				_, err = srv.Client().Database("test").Collection("users").InsertOne(context.TODO(), user2)
				assert.NoError(t, err)
				_, err = srv.Client().Database("test").Collection("namespaces").InsertOne(context.TODO(), namespace1)
				assert.NoError(t, err)
				_, err = srv.Client().Database("test").Collection("namespaces").InsertOne(context.TODO(), namespace2)
				assert.NoError(t, err)
				_, err = srv.Client().Database("test").Collection("namespaces").InsertOne(context.TODO(), namespace3)
				assert.NoError(t, err)

				migrates := migrate.NewMigrate(srv.Client().Database("test"), GenerateMigrations()[48:49]...)
				assert.NoError(t, migrates.Up(context.Background(), migrate.AllAvailable))

				user := new(models.User)
				result := srv.Client().Database("test").Collection("users").FindOne(context.TODO(), bson.M{"_id": user1ID.String()})
				assert.NoError(t, result.Err())

				err = result.Decode(user)
				assert.NoError(t, err)

				assert.Equal(t, 2, user.Namespaces)
			},
		},
		{
			"Success to apply down on migration 49",
			func(t *testing.T) {
				t.Helper()

				user1ID, err := primitive.ObjectIDFromHex("507f1f77bcf86cd799439011")
				assert.NoError(t, err)
				user1 := &models.User{
					ID: user1ID.String(),
				}

				user2ID, err := primitive.ObjectIDFromHex("507f1f77bcf86cd799439012")
				assert.NoError(t, err)
				user2 := &models.User{
					ID: user2ID.String(),
				}

				namespace1 := &models.Namespace{
					Name:  "namespace1",
					Owner: user1ID.String(),
				}
				namespace2 := &models.Namespace{
					Name:  "namespace2",
					Owner: user1ID.String(),
				}
				namespace3 := &models.Namespace{
					Name:  "namespace3",
					Owner: user2ID.String(),
				}

				_, err = srv.Client().Database("test").Collection("users").InsertOne(context.TODO(), user1)
				assert.NoError(t, err)
				_, err = srv.Client().Database("test").Collection("users").InsertOne(context.TODO(), user2)
				assert.NoError(t, err)
				_, err = srv.Client().Database("test").Collection("namespaces").InsertOne(context.TODO(), namespace1)
				assert.NoError(t, err)
				_, err = srv.Client().Database("test").Collection("namespaces").InsertOne(context.TODO(), namespace2)
				assert.NoError(t, err)
				_, err = srv.Client().Database("test").Collection("namespaces").InsertOne(context.TODO(), namespace3)
				assert.NoError(t, err)

				migrates := migrate.NewMigrate(srv.Client().Database("test"), GenerateMigrations()[48:49]...)
				assert.NoError(t, migrates.Down(context.Background(), migrate.AllAvailable))

				user := new(models.User)
				result := srv.Client().Database("test").Collection("users").FindOne(context.TODO(), bson.M{"_id": user1ID.String()})
				assert.NoError(t, result.Err())

				err = result.Decode(user)
				assert.NoError(t, err)

				assert.Equal(t, 0, user.Namespaces)
			},
		},
	}

	for _, test := range cases {
		tc := test
		t.Run(tc.description, func(t *testing.T) {
			t.Cleanup(func() {
				assert.NoError(t, fixtures.Teardown())
			})
			tc.Test(t)
		})
	}
}
