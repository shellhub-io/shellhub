package migrations

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	migrate "github.com/xakep666/mongo-migrate"
)

func TestMigration10(t *testing.T) {
	logrus.Info("Testing Migration 10 - Test if the session_record is not unique")

	db := dbtest.DBServer{}
	defer db.Stop()

	migrates := migrate.NewMigrate(db.Client().Database("test"), GenerateMigrations()[:9]...)
	err := migrates.Up(migrate.AllAvailable)
	assert.NoError(t, err)

	user1 := struct {
		Name          string `json:"name" validate:"required,min=1"`
		Email         string `json:"email" bson:",omitempty" validate:"required,email"`
		Username      string `json:"username" bson:",omitempty" validate:"required,min=3,max=30,alphanum,ascii"`
		Password      string `json:"password" bson:",omitempty"`
		SessionRecord bool   `json:"session_record"`
	}{
		Name:          "user1",
		Email:         "email1",
		Username:      "username1",
		Password:      "password",
		SessionRecord: true,
	}

	user2 := struct {
		Name          string `json:"name" validate:"required,min=1"`
		Email         string `json:"email" bson:",omitempty" validate:"required,email"`
		Username      string `json:"username" bson:",omitempty" validate:"required,min=3,max=30,alphanum,ascii"`
		Password      string `json:"password" bson:",omitempty"`
		SessionRecord bool   `json:"session_record"`
	}{
		Name:          "user2",
		Email:         "email2",
		Username:      "username2",
		Password:      "password",
		SessionRecord: true,
	}

	_, err = db.Client().Database("test").Collection("users").InsertOne(context.TODO(), user1)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("users").InsertOne(context.TODO(), user2)
	assert.NoError(t, err)

	migrates = migrate.NewMigrate(db.Client().Database("test"), GenerateMigrations()[:10]...)
	err = migrates.Up(migrate.AllAvailable)
	assert.NoError(t, err)
}
