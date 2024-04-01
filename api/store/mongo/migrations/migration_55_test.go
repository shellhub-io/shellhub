package migrations

import (
	"context"
	"errors"
	"testing"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	migrate "github.com/xakep666/mongo-migrate"
)

func TestMigration55(t *testing.T) {
	logrus.Info("Testing Migration 55")

	fieldNameTenantID := "tenant_id_1"
	fieldNameTenantIDUID := "tenant_id_1_uid_1"
	fieldNameTimestamp := "timestamp_1"

	db := dbtest.DBServer{}
	defer db.Stop()

	cases := []struct {
		description string
		test        func() error
	}{
		{
			"Success to apply up on migration 55",
			func() error {
				migrations := GenerateMigrations()[54:55]
				migrates := migrate.NewMigrate(db.Client().Database("test"), migrations...)
				err := migrates.Up(context.Background(), migrate.AllAvailable)
				if err != nil {
					return err
				}

				cursor, err := db.Client().Database("test").Collection("removed_devices").Indexes().List(context.Background())
				if err != nil {
					return err
				}

				var foundNameTenantID bool
				var foundNameTenantIDUID bool
				var foundNameTimestamp bool
				for cursor.Next(context.Background()) {
					var index bson.M
					if err := cursor.Decode(&index); err != nil {
						return err
					}

					switch index["name"] {
					case fieldNameTenantID:
						foundNameTenantID = true
					case fieldNameTenantIDUID:
						foundNameTenantIDUID = true
					case fieldNameTimestamp:
						foundNameTimestamp = true
					}
				}

				if !foundNameTenantID || !foundNameTenantIDUID || !foundNameTimestamp {
					return errors.New("one of the indexes was not created")
				}

				return nil
			},
		},
		{
			"Success to apply down on migration 55",
			func() error {
				migrations := GenerateMigrations()[54:55]
				migrates := migrate.NewMigrate(db.Client().Database("test"), migrations...)
				err := migrates.Down(context.Background(), migrate.AllAvailable)
				if err != nil {
					return err
				}

				cursor, err := db.Client().Database("test").Collection("removed_devices").Indexes().List(context.Background())
				if err != nil {
					return errors.New("index not dropped")
				}

				var foundNameTenantID bool
				var foundNameTenantIDUID bool
				var foundNameTimestamp bool
				for cursor.Next(context.Background()) {
					var index bson.M
					if err := cursor.Decode(&index); err != nil {
						return err
					}

					switch index["name"] {
					case fieldNameTenantID:
						foundNameTenantID = true
					case fieldNameTenantIDUID:
						foundNameTenantIDUID = true
					case fieldNameTimestamp:
						foundNameTimestamp = true
					}
				}

				if foundNameTenantID || foundNameTenantIDUID || foundNameTimestamp {
					return errors.New("one of the indexes was deleted")
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
