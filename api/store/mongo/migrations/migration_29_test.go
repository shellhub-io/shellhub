package migrations

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestMigration29(t *testing.T) {
	logrus.Info("Testing Migration 29 - Test whether the collection of users the field last_login was created")

	db := dbtest.DBServer{}
	defer db.Stop()

	user := models.User{
		Name: "Test",
	}

	_, err := db.Client().Database("test").Collection("users").InsertOne(context.TODO(), user)
	assert.NoError(t, err)

	migrations := GenerateMigrations()[:29]

	migrates := migrate.NewMigrate(db.Client().Database("test"), migrations...)
	err = migrates.Up(migrate.AllAvailable)
	assert.NoError(t, err)

	version, _, err := migrates.Version()
	assert.NoError(t, err)
	assert.Equal(t, uint64(29), version)

	var migratedUser *models.User
	err = db.Client().Database("test").Collection("users").FindOne(context.TODO(), bson.M{"name": user.Name}).Decode(&migratedUser)
	assert.NoError(t, err)

	index := db.Client().Database("test").Collection("users").Indexes()

	cursor, err := index.List(context.TODO())
	assert.NoError(t, err)

	var results []bson.M
	err = cursor.All(context.TODO(), &results)
	assert.NoError(t, err)

	keyField := results[1]["key"].(primitive.M)
	assert.Equal(t, nil, keyField["last_login"])
}
