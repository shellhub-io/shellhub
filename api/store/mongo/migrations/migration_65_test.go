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

func TestMigration65(t *testing.T) {
	ctx := context.Background()

	mock := &envMocks.Backend{}
	envs.DefaultBackend = mock

	cases := []struct {
		description string
		setup       func() error
		test        func() error
	}{
		{
			description: "Success to apply up on migration 65",
			setup: func() error {
				_, err := c.
					Database("test").
					Collection("users").
					InsertOne(ctx, models.User{
						UserData: models.UserData{
							Username: "john_doe",
						},
					})

				return err
			},
			test: func() error {
				migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[64])
				if err := migrates.Up(context.Background(), migrate.AllAvailable); err != nil {
					return err
				}

				query := c.
					Database("test").
					Collection("users").
					FindOne(context.TODO(), bson.M{"username": "john_doe"})

				user := new(models.User)
				if err := query.Decode(user); err != nil {
					return errors.New("unable to find the user")
				}

				if user.RecoveryEmail != "" {
					return errors.New("unable to apply the migration")
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
