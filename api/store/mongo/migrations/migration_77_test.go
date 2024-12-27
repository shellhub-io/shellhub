package migrations

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/envs"
	envmock "github.com/shellhub-io/shellhub/pkg/envs/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	migrate "github.com/xakep666/mongo-migrate"
)

func TestMigration77Up(t *testing.T) {
	ctx := context.Background()

	mock := &envmock.Backend{}
	envs.DefaultBackend = mock

	cases := []struct {
		setup       func() error
		description string
	}{
		{
			description: "Success to apply up on migration 77",
			setup: func() error {
				_, err := c.
					Database("test").
					Collection("users").
					InsertOne(ctx, map[string]interface{}{
						"username": nil,
					})

				return err
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.description, func(tt *testing.T) {
			tt.Cleanup(func() {
				assert.NoError(tt, srv.Reset())
			})

			assert.NoError(tt, tc.setup())

			migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[76])
			require.NoError(tt, migrates.Up(context.Background(), migrate.AllAvailable))

			_, err := c.
				Database("test").
				Collection("users").
				InsertOne(ctx, map[string]interface{}{
					"username": nil,
				})
			require.NoError(tt, err)
		})
	}
}
