package migrations

import (
	"context"
	"errors"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/envs"
	envMocks "github.com/shellhub-io/shellhub/pkg/envs/mocks"
	"github.com/stretchr/testify/assert"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
)

func TestMigration62Up(t *testing.T) {
	cases := []struct {
		description string
		mocks       func()
		expected    func() error
	}{
		{
			description: "Success to apply up on migration 62",
			mocks: func() {
				mock := &envMocks.Backend{}
				envs.DefaultBackend = mock
				mock.On("Get", "SHELLHUB_CLOUD").Return("true").Once()
			},
			expected: func() error {
				cursor, err := c.Database("test").Collection("recorded_sessions").Indexes().List(context.Background())
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
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			t.Cleanup(func() {
				assert.NoError(t, srv.Reset())
			})

			tc.mocks()

			migrations := GenerateMigrations()[61:62]
			migrates := migrate.NewMigrate(c.Database("test"), migrations...)
			assert.NoError(t, migrates.Up(context.Background(), migrate.AllAvailable))

			assert.NoError(t, tc.expected())
		})
	}
}

func TestMigration62Down(t *testing.T) {
	mock := &envMocks.Backend{}
	envs.DefaultBackend = mock

	cases := []struct {
		description string
		mocks       func()
		expected    func() error
	}{
		{
			description: "Success to apply down on migration 62",
			mocks:       func() {},
			expected: func() error {
				cursor, err := c.Database("test").Collection("recorded_sessions").Indexes().List(context.Background())
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

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			t.Cleanup(func() {
				assert.NoError(t, srv.Reset())
			})

			tc.mocks()

			migrations := GenerateMigrations()[61:62]
			migrates := migrate.NewMigrate(c.Database("test"), migrations...)
			assert.NoError(t, migrates.Down(context.Background(), migrate.AllAvailable))

			assert.NoError(t, tc.expected())
		})
	}
}
