package migrations

import (
	"context"
	"errors"
	"testing"

	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	"github.com/shellhub-io/shellhub/pkg/envs"
	envMocks "github.com/shellhub-io/shellhub/pkg/envs/mocks"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
)

func TestMigration64(t *testing.T) {
	logrus.Info("Testing Migration 64")

	ctx := context.Background()

	db := dbtest.DBServer{}
	defer db.Stop()

	mock := &envMocks.Backend{}
	envs.DefaultBackend = mock

	cases := []struct {
		description string
		setup       func() error
		test        func() error
	}{
		{
			description: "Success to apply up on migration 64",
			setup: func() error {
				_, err := db.
					Client().
					Database("test").
					Collection("namespaces").
					InsertOne(ctx, models.Namespace{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Settings: &models.NamespaceSettings{},
					})

				return err
			},
			test: func() error {
				migrations := GenerateMigrations()[63:64]
				migrates := migrate.NewMigrate(db.Client().Database("test"), migrations...)
				err := migrates.Up(context.Background(), migrate.AllAvailable)
				if err != nil {
					return err
				}

				query := db.
					Client().
					Database("test").
					Collection("namespaces").
					FindOne(context.TODO(), bson.M{"tenant_id": "00000000-0000-4000-0000-000000000000"})

				ns := new(models.Namespace)
				if err := query.Decode(ns); err != nil {
					return errors.New("unable to find the namespace")
				}

				if ns.Settings.ConnectionAnnouncement != "" {
					return errors.New("unable to apply the migration")
				}

				return nil
			},
		},
		{
			description: "Success to apply down on migration 64",
			setup: func() error {
				_, err := db.
					Client().
					Database("test").
					Collection("namespaces").
					InsertOne(ctx, models.Namespace{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Settings: &models.NamespaceSettings{
							ConnectionAnnouncement: "My awesome connection announcement",
						},
					})

				return err
			},
			test: func() error {
				migrations := GenerateMigrations()[63:64]
				migrates := migrate.NewMigrate(db.Client().Database("test"), migrations...)
				err := migrates.Down(context.Background(), migrate.AllAvailable)
				if err != nil {
					return err
				}

				query := db.
					Client().
					Database("test").
					Collection("namespaces").
					FindOne(context.TODO(), bson.M{"tenant_id": "00000000-0000-4000-0000-000000000000"})

				ns := new(models.Namespace)
				if err := query.Decode(ns); err != nil {
					return errors.New("unable to find the namespace")
				}

				if ns.Settings.ConnectionAnnouncement != "" {
					return errors.New("unable to apply the migration")
				}

				return nil
			},
		},
	}

	for _, test := range cases {
		tc := test
		t.Run(tc.description, func(t *testing.T) {
			err := tc.setup()
			assert.NoError(t, err)

			err = tc.test()
			assert.NoError(t, err)
		})
	}
}
