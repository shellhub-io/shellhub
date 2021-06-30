package migrations

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	migrate "github.com/xakep666/mongo-migrate"
)

func TestMigration5(t *testing.T) {
	logrus.Info("Testing Migration 5 - Test if the email is set unique")

	db := dbtest.DBServer{}
	defer db.Stop()

	user1 := models.User{Name: "name1", Username: "username1", Password: "password", Email: "email"}
	user2 := models.User{Name: "name2", Username: "username2", Password: "password", Email: "email"}

	_, err := db.Client().Database("test").Collection("users").InsertOne(context.TODO(), user1)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("users").InsertOne(context.TODO(), user2)
	assert.NoError(t, err)

	migrates := migrate.NewMigrate(db.Client().Database("test"), GenerateMigrations()[:4]...)
	err = migrates.Up(migrate.AllAvailable)
	assert.NoError(t, err)

	migrates = migrate.NewMigrate(db.Client().Database("test"), GenerateMigrations()[:5]...)
	err = migrates.Up(migrate.AllAvailable)
	assert.Error(t, err)
}
