package migrations

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/envs"
	envmock "github.com/shellhub-io/shellhub/pkg/envs/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
)

func TestMigration93Up(t *testing.T) {
	ctx := context.Background()
	mock := &envmock.Backend{}
	envs.DefaultBackend = mock

	cases := []struct {
		description string
		setup       func() error
		verify      func(tt *testing.T)
	}{
		{
			description: "removes the public_url attribute from devices",
			setup: func() error {
				_, err := c.Database("test").Collection("devices").InsertOne(ctx, bson.M{"uid": "uid", "public_url": true})

				return err
			},
			verify: func(tt *testing.T) {
				device := make(map[string]any)
				require.NoError(tt, c.Database("test").Collection("devices").FindOne(ctx, bson.M{"uid": "uid"}).Decode(&device))

				_, ok := device["public_url"]
				require.Equal(tt, false, ok)
			},
		},
		{
			description: "removes the public_url_address attribute from devices",
			setup: func() error {
				_, err := c.Database("test").Collection("devices").InsertOne(ctx, bson.M{"uid": "uid", "public_url_address": "address"})

				return err
			},
			verify: func(tt *testing.T) {
				device := make(map[string]any)
				require.NoError(tt, c.Database("test").Collection("devices").FindOne(ctx, bson.M{"uid": "uid"}).Decode(&device))

				_, ok := device["public_url_address"]
				require.Equal(tt, false, ok)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			tt.Cleanup(func() {
				assert.NoError(tt, srv.Reset())
			})

			require.NoError(tt, tc.setup())

			migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[92])
			require.NoError(tt, migrates.Up(ctx, migrate.AllAvailable))

			tc.verify(tt)
		})
	}
}
