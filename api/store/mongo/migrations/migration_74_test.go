package migrations

import (
	"context"
	"errors"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/envs"
	envMocks "github.com/shellhub-io/shellhub/pkg/envs/mocks"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
)

func TestMigration74(t *testing.T) {
	ctx := context.Background()

	mock := &envMocks.Backend{}
	envs.DefaultBackend = mock

	cases := []struct {
		description string
		setup       func() error
		test        func() error
	}{
		{
			description: "Success to apply up on migration 74",
			setup: func() error {
				_, err := c.
					Database("test").
					Collection("namespaces").
					InsertOne(ctx, models.Namespace{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Settings: &models.NamespaceSettings{},
					})

				return err
			},
			test: func() error {
				migrations := GenerateMigrations()[73:74]
				migrates := migrate.NewMigrate(c.Database("test"), migrations...)
				err := migrates.Up(context.Background(), migrate.AllAvailable)
				if err != nil {
					return err
				}

				query := c.
					Database("test").
					Collection("namespaces").
					FindOne(context.TODO(), bson.M{"tenant_id": "00000000-0000-4000-0000-000000000000"})

				ns := new(models.Namespace)
				if err := query.Decode(ns); err != nil {
					return errors.New("unable to find the namespace")
				}

				if ns.Settings.ConnectionAnnouncement != defaultAnnouncementMessage {
					return errors.New("unable to apply the migration")
				}

				return nil
			},
		}, {
			description: "Success to unapply the migration 74",
			setup: func() error {
				_, err := c.
					Database("test").
					Collection("namespaces").
					InsertOne(ctx, models.Namespace{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Settings: &models.NamespaceSettings{},
					})

				return err
			},
			test: func() error {
				migrations := GenerateMigrations()[73:74]
				migrates := migrate.NewMigrate(c.Database("test"), migrations...)
				err := migrates.Down(context.Background(), migrate.AllAvailable)
				if err != nil {
					return err
				}

				query := c.
					Database("test").
					Collection("namespaces").
					FindOne(context.TODO(), bson.M{"tenant_id": "00000000-0000-4000-0000-000000000000"})

				ns := new(models.Namespace)
				if err := query.Decode(ns); err != nil {
					return errors.New("unable to find the namespace")
				}

				if ns.Settings.ConnectionAnnouncement != "" {
					return errors.New("unable to unapply the migration")
				}

				return nil
			},
		},
	}

	for _, test := range cases {
		tc := test
		t.Run(tc.description, func(t *testing.T) {
			t.Cleanup(func() {
				assert.NoError(t, srv.Reset())
			})

			assert.NoError(t, tc.setup())
			assert.NoError(t, tc.test())
		})
	}
}
