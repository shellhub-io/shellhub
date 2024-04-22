package migrations

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/envs"
	envMocks "github.com/shellhub-io/shellhub/pkg/envs/mocks"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/require"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
)

func TestMigration65(t *testing.T) {
	mock := &envMocks.Backend{}
	envs.DefaultBackend = mock

	cases := []struct {
		description string
		setup       func(context.Context) error
		test        func(context.Context) error
	}{
		{
			description: "Success to apply up on migration 65",
			setup: func(ctx context.Context) error {
				_, err := c.
					Database("test").
					Collection("devices").
					InsertOne(ctx, models.Device{
						UID: "0000000000000000000000000000000000000000000000000000000000000000",
					})

				return err
			},
			test: func(ctx context.Context) error {
				migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[64])

				if err := migrates.Up(ctx, migrate.AllAvailable); err != nil {
					return err
				}

				query := c.
					Database("test").
					Collection("devices").
					FindOne(ctx, bson.M{"uid": "0000000000000000000000000000000000000000000000000000000000000000"})

				dev := new(models.Device)
				if err := query.Decode(dev); err != nil {
					return errors.New("unable to find the device")
				}

				z := time.Time{} // zero value
				if dev.ConnectedAt != z && dev.DisconnectedAt != z {
					return errors.New("unable to apply the migration")
				}

				return nil
			},
		},
		{
			description: "Success to apply down on migration 64",
			setup: func(ctx context.Context) error {
				_, err := c.
					Database("test").
					Collection("devices").
					InsertOne(ctx, models.Device{
						UID:         "0000000000000000000000000000000000000000000000000000000000000000",
						ConnectedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
					})

				return err
			},
			test: func(ctx context.Context) error {
				migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[64])

				if err := migrates.Down(ctx, migrate.AllAvailable); err != nil {
					return err
				}

				query := c.
					Database("test").
					Collection("devices").
					FindOne(ctx, bson.M{"uid": "0000000000000000000000000000000000000000000000000000000000000000"})

				dev := new(models.Device)
				if err := query.Decode(dev); err != nil {
					return errors.New("unable to find the device")
				}

				z := time.Time{} // zero value
				if dev.ConnectedAt != z && dev.DisconnectedAt != z {
					return errors.New("unable to apply the migration")
				}

				return nil
			},
		},
	}

	for _, test := range cases {
		tc := test
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()
			require.NoError(t, tc.setup(ctx))
			require.NoError(t, tc.test(ctx))
		})
	}
}
