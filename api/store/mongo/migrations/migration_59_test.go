package migrations

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
)

func TestMigration59(t *testing.T) {
	logrus.Info("Testing Migration 59")

	ctx := context.TODO()
	db := dbtest.DB{}
	err := func() error {
		err := db.Down(context.Background())

		return err
	}()
	assert.NoError(t, err)

	type Expected struct {
		user *models.User
		err  error
	}

	cases := []struct {
		description string
		setup       func() (func() error, error)
		check       func() (*models.User, error)
		expected    Expected
	}{
		{
			description: "Success to apply up on migration 59",
			setup: func() (func() error, error) {
				if _, err = mongoClient.Database("test").Collection("users").InsertOne(ctx, models.User{
					ID:        "652594bcc7b001c6f298df48",
					CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
					LastLogin: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
					UserData: models.UserData{
						Name:     "John Doe",
						Email:    "JohnDoe@test.com",
						Username: "John Doe",
					},
					Password: models.UserPassword{
						Hash: "2bb80d537b1da3e38bd30361aa855686bde0eacd7162fef6a25fe97bf527a25b",
					},
				}); err != nil {
					return nil, err
				}

				user := new(models.User)
				if err := mongoClient.Database("test").Collection("users").FindOne(ctx, bson.M{"name": "John Doe"}).Decode(&user); err != nil {
					return nil, err
				}

				return func() error {
					d, err := mongoClient.Database("test").Collection("users").DeleteOne(ctx, bson.M{"username": "john doe"})
					if err != nil {
						return err
					}

					if d.DeletedCount < 1 {
						return errors.New("No users deleted")
					}

					return nil
				}, nil
			},
			check: func() (*models.User, error) {
				user := new(models.User)

				if err := mongoClient.Database("test").Collection("users").FindOne(ctx, bson.M{"username": "john doe"}).Decode(&user); err != nil {
					return nil, err
				}

				return user, nil
			},
			expected: Expected{
				user: &models.User{
					ID:             "652594bcc7b001c6f298df48",
					Namespaces:     0,
					MaxNamespaces:  0,
					Confirmed:      false,
					CreatedAt:      time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
					LastLogin:      time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
					EmailMarketing: false,
					UserData: models.UserData{
						Name:     "John Doe",
						Email:    "johndoe@test.com",
						Username: "john doe",
					},
					Password: models.UserPassword{
						Hash: "2bb80d537b1da3e38bd30361aa855686bde0eacd7162fef6a25fe97bf527a25b",
					},
				},
				err: nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			teardown, err := tc.setup()
			assert.NoError(t, err)

			migrates := migrate.NewMigrate(mongoClient.Database("test"), migration59)
			assert.NoError(t, migrates.Up(context.Background(), migrate.AllAvailable))

			user, err := tc.check()
			assert.Equal(t, tc.expected, Expected{user, err})

			assert.NoError(t, teardown())
		})
	}
}
