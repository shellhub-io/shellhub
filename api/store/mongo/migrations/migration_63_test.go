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
)

func TestMigration63(t *testing.T) {
	logrus.Info("Testing Migration 63 - Test whether MFA fields were added to the users collection")

	db := dbtest.DB{}
	defer db.Stop()

	user := models.User{
		UserData: models.UserData{
			Name: "Test",
		},
	}

	_, err := db.Client().Database("test").Collection("users").InsertOne(context.TODO(), user)
	assert.NoError(t, err)

	migrations := GenerateMigrations()[62:63]

	migrates := migrate.NewMigrate(db.Client().Database("test"), migrations...)
	err = migrates.Up(context.Background(), migrate.AllAvailable)
	assert.NoError(t, err)

	version, _, err := migrates.Version(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, uint64(63), version)

	var migratedUser *models.User
	err = db.Client().Database("test").Collection("users").FindOne(context.TODO(), bson.M{"name": user.Name}).Decode(&migratedUser)
	assert.NoError(t, err)
	assert.False(t, migratedUser.MFA)
	assert.Equal(t, "", migratedUser.Secret)
	assert.Empty(t, migratedUser.Codes)

	err = migrates.Down(context.Background(), migrate.AllAvailable)
	assert.NoError(t, err)

	err = db.Client().Database("test").Collection("users").FindOne(context.TODO(), bson.M{"name": user.Name}).Decode(&migratedUser)
	assert.NoError(t, err)
	assert.False(t, migratedUser.MFA)
	assert.Equal(t, "", migratedUser.Secret)
	assert.Empty(t, migratedUser.Codes)
}
