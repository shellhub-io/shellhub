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

	db := dbtest.DB{}
	err := func() error {
		err := db.Down(context.Background())

		return err
	}()
	assert.NoError(t, err)

	user1 := models.User{
		UserData: models.UserData{
			Name:     "name1",
			Username: "username1",
			Email:    "email",
		},
		Password: models.UserPassword{
			Hash: "2bb80d537b1da3e38bd30361aa855686bde0eacd7162fef6a25fe97bf527a25b",
		},
	}
	user2 := models.User{
		UserData: models.UserData{
			Name:     "name2",
			Username: "username2",
			Email:    "email",
		},
		Password: models.UserPassword{
			Hash: "2bb80d537b1da3e38bd30361aa855686bde0eacd7162fef6a25fe97bf527a25b",
		},
	}

	_, err = mongoClient.Database("test").Collection("users").InsertOne(context.TODO(), user1)
	assert.NoError(t, err)

	_, err = mongoClient.Database("test").Collection("users").InsertOne(context.TODO(), user2)
	assert.NoError(t, err)

	migrates := migrate.NewMigrate(mongoClient.Database("test"), GenerateMigrations()[:4]...)
	err = migrates.Up(context.Background(), migrate.AllAvailable)
	assert.NoError(t, err)

	migrates = migrate.NewMigrate(mongoClient.Database("test"), GenerateMigrations()[:5]...)
	err = migrates.Up(context.Background(), migrate.AllAvailable)
	assert.Error(t, err)
}
