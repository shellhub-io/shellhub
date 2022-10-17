package migrations

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestMigration50(t *testing.T) {
	logrus.Info("Testing Migration 50")

	var maxNamespacesWanted int
	if envs.IsCloud() {
		maxNamespacesWanted = 1
	} else {
		maxNamespacesWanted = -1
	}

	db := dbtest.DBServer{}
	defer db.Stop()

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

	_, err = db.Client().Database("test").Collection("users").InsertOne(context.TODO(), user1)
	assert.NoError(t, err)
	_, err = db.Client().Database("test").Collection("users").InsertOne(context.TODO(), user2)
	assert.NoError(t, err)
	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(context.TODO(), namespace1)
	assert.NoError(t, err)
	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(context.TODO(), namespace2)
	assert.NoError(t, err)
	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(context.TODO(), namespace3)
	assert.NoError(t, err)

	cases := []struct {
		description string
		Test        func(t *testing.T)
	}{
		{
			"Success to apply up on migration 50",
			func(t *testing.T) {
				t.Helper()

				migrations := GenerateMigrations()[48:50]
				migrates := migrate.NewMigrate(db.Client().Database("test"), migrations...)
				err := migrates.Up(migrate.AllAvailable)
				assert.NoError(t, err)

				user := new(models.User)
				result := db.Client().Database("test").Collection("users").FindOne(context.TODO(), bson.M{"_id": user1ID.String()})
				assert.NoError(t, result.Err())

				err = result.Decode(user)
				assert.NoError(t, err)

				assert.Equal(t, maxNamespacesWanted, user.MaxNamespaces)
			},
		},
		{
			"Success to apply down on migration 50",
			func(t *testing.T) {
				t.Helper()

				migrations := GenerateMigrations()[48:50]
				migrates := migrate.NewMigrate(db.Client().Database("test"), migrations...)
				err := migrates.Down(migrate.AllAvailable)
				assert.NoError(t, err)

				user := new(models.User)
				result := db.Client().Database("test").Collection("users").FindOne(context.TODO(), bson.M{"_id": user1ID.String()})
				assert.NoError(t, result.Err())

				err = result.Decode(user)
				assert.NoError(t, err)

				assert.Equal(t, 0, user.MaxNamespaces)
			},
		},
	}

	for _, test := range cases {
		tc := test
		t.Run(tc.description, tc.Test)
	}
}
