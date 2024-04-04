package migrations

import (
	"context"
	"errors"
	"testing"

	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
)

func TestMigration54(t *testing.T) {
	logrus.Info("Testing Migration 54")

	const Name string = "tenant_id_1_status_1"

	db := dbtest.DB{}
	defer db.Stop()

	cases := []struct {
		description string
		test        func() error
	}{
		{
			"Success to apply up on migration 54",
			func() error {
				migrations := GenerateMigrations()[53:54]
				migrates := migrate.NewMigrate(db.Client().Database("test"), migrations...)
				err := migrates.Up(context.Background(), migrate.AllAvailable)
				if err != nil {
					return err
				}

				cursor, err := db.Client().Database("test").Collection("devices").Indexes().List(context.Background())
				if err != nil {
					return err
				}

				var found bool
				for cursor.Next(context.Background()) {
					var index bson.M
					if err := cursor.Decode(&index); err != nil {
						return err
					}

					if index["name"] == Name {
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
			"Success to apply down on migration 54",
			func() error {
				migrations := GenerateMigrations()[53:54]
				migrates := migrate.NewMigrate(db.Client().Database("test"), migrations...)
				err := migrates.Down(context.Background(), migrate.AllAvailable)
				if err != nil {
					return err
				}

				cursor, err := db.Client().Database("test").Collection("devices").Indexes().List(context.Background())
				if err != nil {
					return errors.New("index not dropped")
				}

				var found bool
				for cursor.Next(context.Background()) {
					var index bson.M
					if err := cursor.Decode(&index); err != nil {
						return err
					}

					if index["name"] == Name {
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
