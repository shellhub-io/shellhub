package migrations

import (
	"context"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
)

func TestMigration38(t *testing.T) {
	logrus.Info("Testing Migration 38")

	db := dbtest.DBServer{}
	defer db.Stop()

	type Expected struct {
		CreatedAt string
		LastLogin string
	}

	migrations := GenerateMigrations()[37:38]

	timeZero := time.Time{}
	timePast := time.Date(2010, 1, 1, 1, 1, 1, 1, time.UTC)
	timeNow := time.Now().UTC()
	convertDate := func(t time.Time) string {
		return t.Format("2006-01-02 15:04:05.000")
	}

	userNoCreatedAt := models.User{
		ID:         "userNoCreatedID",
		Namespaces: 0,
		Confirmed:  false,
		CreatedAt:  timeZero,
		LastLogin:  timeNow,
		UserData: models.UserData{
			Name:     "userNoCreatedAt",
			Email:    "userNoCreatedAt@mail.com",
			Username: "userNoCreatedAt",
		},
		UserPassword: models.UserPassword{
			Password: "",
		},
	}
	userWithCreatedAt := models.User{
		ID:         "userWithCreatedID",
		Namespaces: 0,
		Confirmed:  false,
		CreatedAt:  timePast,
		LastLogin:  timeNow,
		UserData: models.UserData{
			Name:     "userWithCreatedAt",
			Email:    "userWithCreatedAt@mail.com",
			Username: "userWithCreatedAt",
		},
		UserPassword: models.UserPassword{
			Password: "",
		},
	}

	_, err := db.Client().Database("test").Collection("users").InsertOne(context.TODO(), userNoCreatedAt)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("users").InsertOne(context.TODO(), userWithCreatedAt)
	assert.NoError(t, err)

	migrates := migrate.NewMigrate(db.Client().Database("test"), migrations...)
	err = migrates.Up(migrate.AllAvailable)
	assert.NoError(t, err)

	cases := []struct {
		description string
		Test        func(t *testing.T)
	}{
		{
			description: "Executes migration when user's created_at property is empty",
			Test: func(t *testing.T) {
				t.Helper()

				var userMigrated *models.User
				err = db.Client().Database("test").Collection("users").FindOne(context.TODO(), bson.D{{"username", userNoCreatedAt.Username}}).Decode(&userMigrated)
				assert.NoError(t, err)
				assert.Equal(t,
					Expected{CreatedAt: convertDate(userNoCreatedAt.LastLogin), LastLogin: convertDate(userNoCreatedAt.LastLogin)},
					Expected{CreatedAt: convertDate(userMigrated.CreatedAt), LastLogin: convertDate(userMigrated.LastLogin)},
				)
			},
		},
		{
			description: "Does not execute migration when user's created_at is already set",
			Test: func(t *testing.T) {
				t.Helper()

				var userMigrated *models.User
				err = db.Client().Database("test").Collection("users").FindOne(context.TODO(), bson.D{{"username", userWithCreatedAt.Username}}).Decode(&userMigrated)
				assert.NoError(t, err)
				assert.Equal(t,
					Expected{CreatedAt: convertDate(userWithCreatedAt.CreatedAt), LastLogin: convertDate(userWithCreatedAt.LastLogin)},
					Expected{CreatedAt: convertDate(userMigrated.CreatedAt), LastLogin: convertDate(userMigrated.LastLogin)},
				)
			},
		},
	}

	for _, test := range cases {
		tc := test
		t.Run(tc.description, tc.Test)
	}
}
