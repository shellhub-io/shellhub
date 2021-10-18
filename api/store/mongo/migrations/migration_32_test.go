package migrations

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
)

func TestMigration32(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	migrations := GenerateMigrations()[:31]

	migrates := migrate.NewMigrate(db.Client().Database("test"), migrations...)
	err := migrates.Up(migrate.AllAvailable)
	assert.NoError(t, err)

	version, _, err := migrates.Version()
	assert.NoError(t, err)
	assert.Equal(t, uint64(31), version)

	user := models.User{
		UserData: models.UserData{
			Name:     "name",
			Username: "username",
			Email:    "email@mail.com",
		},
		UserPassword: models.UserPassword{
			Password: "password",
		},
	}
	_, err = db.Client().Database("test").Collection("users").InsertOne(context.TODO(), user)
	assert.NoError(t, err)

	migration := GenerateMigrations()[31]

	migrates = migrate.NewMigrate(db.Client().Database("test"), migration)
	err = migrates.Up(migrate.AllAvailable)
	assert.NoError(t, err)

	version, _, err = migrates.Version()
	assert.NoError(t, err)
	assert.Equal(t, uint64(32), version)

	var migratedUser *models.User
	err = db.Client().Database("test").Collection("users").FindOne(context.TODO(), bson.M{"name": "name"}).Decode(&migratedUser)
	assert.NoError(t, err)
	assert.Equal(t, true, migratedUser.Authenticated)
}
