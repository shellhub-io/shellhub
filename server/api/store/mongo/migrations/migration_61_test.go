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

func TestMigration61(t *testing.T) {
	ctx := context.Background()

	mock := &envMocks.Backend{}
	envs.DefaultBackend = mock

	cases := []struct {
		description string
		setup       func() (func() error, error)
		test        func() error
	}{
		{
			"Success to apply up on migration 61",
			func() (func() error, error) {
				if _, err := c.Database("test").Collection("devices").InsertOne(ctx, models.Device{
					Name: "",
				}); err != nil {
					return nil, err
				}

				if _, err := c.Database("test").Collection("devices").InsertOne(ctx, models.Device{
					Name: "test",
				}); err != nil {
					return nil, err
				}

				return func() error {
					_, err := c.Database("test").Collection("devices").DeleteOne(ctx, bson.M{
						"name": "test",
					})
					if err != nil {
						return err
					}

					return nil
				}, nil
			},
			func() error {
				migrations := GenerateMigrations()[60:61]
				migrates := migrate.NewMigrate(c.Database("test"), migrations...)
				err := migrates.Up(context.Background(), migrate.AllAvailable)
				if err != nil {
					return err
				}

				count, err := c.Database("test").Collection("devices").CountDocuments(ctx, bson.M{"name": ""})
				if err != nil {
					return err
				}

				if count != 0 {
					return errors.New("failed because don't deleted the expected")
				}

				count, err = c.Database("test").Collection("devices").CountDocuments(ctx, bson.M{"name": "test"})
				if err != nil {
					return err
				}

				if count != 1 {
					return errors.New("failed because deleted more than the expected")
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

			teardown, err := tc.setup()
			assert.NoError(t, err)

			err = tc.test()
			assert.NoError(t, err)

			if teardown != nil {
				assert.NoError(t, teardown())
			}
		})
	}
}
