package migrations

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/envs"
	envmock "github.com/shellhub-io/shellhub/pkg/envs/mocks"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
)

func TestMigration79(t *testing.T) {
	ctx := context.Background()

	mock := &envmock.Backend{}
	envs.DefaultBackend = mock

	mock.On("Get", "SHELLHUB_CLOUD").Return("false")
	mock.On("Get", "SHELLHUB_ENTERPRISE").Return("false")

	tests := []struct {
		description string
		setup       func(t *testing.T)
		run         func(t *testing.T)
	}{
		{
			description: "Apply up on migration 79 when there is no user",
			setup:       func(_ *testing.T) {},
			run: func(t *testing.T) {
				result := c.Database("test").Collection("system").FindOne(ctx, bson.M{})
				require.NoError(t, result.Err())

				var system models.System

				err := result.Decode(&system)
				require.NoError(t, err)

				assert.Equal(t, false, system.Setup)
			},
		},
		{
			description: "Apply up on migration 79 when there is at least one user",
			setup: func(t *testing.T) {
				_, err := c.Database("test").Collection("users").InsertOne(ctx, models.User{})
				require.NoError(t, err)
			},
			run: func(t *testing.T) {
				result := c.Database("test").Collection("system").FindOne(ctx, bson.M{})
				require.NoError(t, result.Err())

				var system models.System

				err := result.Decode(&system)
				require.NoError(t, err)

				assert.Equal(t, true, system.Setup)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(tt *testing.T) {
			tt.Cleanup(func() {
				assert.NoError(tt, srv.Reset())
			})

			test.setup(tt)

			migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[78])
			require.NoError(tt, migrates.Up(context.Background(), migrate.AllAvailable))

			test.run(tt)
		})
	}
}
