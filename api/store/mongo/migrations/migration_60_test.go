package migrations

import (
	"context"
	"errors"
	"testing"

	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	"github.com/shellhub-io/shellhub/pkg/envs"
	envMocks "github.com/shellhub-io/shellhub/pkg/envs/mocks"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
)

func TestMigration60(t *testing.T) {
	logrus.Info("Testing Migration 60")

	db := dbtest.DBServer{}
	defer db.Stop()

	mock := &envMocks.Backend{}
	envs.DefaultBackend = mock

	cases := []struct {
		description string
		test        func() error
	}{
		{
			"Success to apply up on migration 60",
			func() error {
				mock.On("Get", "SHELLHUB_CLOUD").Return("true").Once()

				migrations := GenerateMigrations()[59:60]
				migrates := migrate.NewMigrate(db.Client().Database("test"), migrations...)
				err := migrates.Up(migrate.AllAvailable)
				if err != nil {
					return err
				}

				cursor, err := db.Client().Database("test").Collection("active_sessions").Indexes().List(context.Background())
				if err != nil {
					return err
				}

				var found bool
				for cursor.Next(context.Background()) {
					var index bson.M
					if err := cursor.Decode(&index); err != nil {
						return err
					}

					if index["name"] == "tenant_id" {
						found = true
					}
				}

				if !found {
					return errors.New("index not created")
				}

				return nil
			},
		},
		{
			"Success to apply down on migration 60",
			func() error {
				migrations := GenerateMigrations()[59:60]
				migrates := migrate.NewMigrate(db.Client().Database("test"), migrations...)
				err := migrates.Down(migrate.AllAvailable)
				if err != nil {
					return err
				}

				cursor, err := db.Client().Database("test").Collection("active_sessions").Indexes().List(context.Background())
				if err != nil {
					return errors.New("index not dropped")
				}

				var found bool
				for cursor.Next(context.Background()) {
					var index bson.M
					if err := cursor.Decode(&index); err != nil {
						return err
					}

					if index["name"] == "tenant_id" {
						found = true
					}
				}

				if found {
					return errors.New("index not dropped")
				}

				return nil
			},
		},
	}

	for _, test := range cases {
		tc := test
		t.Run(tc.description, func(t *testing.T) {
			err := tc.test()
			assert.NoError(t, err)
		})
	}
}
