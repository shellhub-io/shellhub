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
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestMigration72Up(t *testing.T) {
	ctx := context.Background()

	mock := &envmock.Backend{}
	envs.DefaultBackend = mock

	cases := []struct {
		description string
		setup       func() error
		test        func() error
	}{
		{
			description: "Success to apply up on migration 72",
			setup: func() error {
				_, err := c.
					Database("test").
					Collection("namespaces").
					InsertOne(ctx, map[string]interface{}{
						"tenant_id": "00000000-0000-4000-0000-000000000000",
						"members": []map[string]interface{}{
							{
								"id": "000000000000000000000000",
							},
						},
					})

				return err
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.description, func(t *testing.T) {
			t.Cleanup(func() {
				assert.NoError(t, srv.Reset())
			})

			assert.NoError(t, tc.setup())

			migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[71])
			require.NoError(t, migrates.Up(context.Background(), migrate.AllAvailable))

			query := c.
				Database("test").
				Collection("namespaces").
				FindOne(context.TODO(), bson.M{"tenant_id": "00000000-0000-4000-0000-000000000000"})

			namespace := make(map[string]interface{})
			require.NoError(t, query.Decode(&namespace))

			members := namespace["members"].(primitive.A)
			for _, m := range members {
				val, ok := m.(map[string]interface{})["status"]
				require.Equal(t, true, ok)
				require.Equal(t, "accepted", val)
			}
		})
	}
}

func TestMigration72Down(t *testing.T) {
	ctx := context.Background()

	mock := &envmock.Backend{}
	envs.DefaultBackend = mock

	cases := []struct {
		description string
		setup       func() error
		test        func() error
	}{
		{
			description: "Success to apply up on migration 72",
			setup: func() error {
				_, err := c.
					Database("test").
					Collection("namespaces").
					InsertOne(ctx, map[string]interface{}{
						"tenant_id": "00000000-0000-4000-0000-000000000000",
						"members": []map[string]interface{}{
							{
								"id":     "000000000000000000000000",
								"status": "accepted",
							},
						},
					})

				return err
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.description, func(t *testing.T) {
			t.Cleanup(func() {
				assert.NoError(t, srv.Reset())
			})

			assert.NoError(t, tc.setup())

			migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[71])
			require.NoError(t, migrates.Up(context.Background(), migrate.AllAvailable))
			require.NoError(t, migrates.Down(context.Background(), migrate.AllAvailable))

			query := c.
				Database("test").
				Collection("namespaces").
				FindOne(context.TODO(), bson.M{"tenant_id": "00000000-0000-4000-0000-000000000000"})

			namespace := make(map[string]interface{})
			require.NoError(t, query.Decode(&namespace))

			members := namespace["members"].(primitive.A)
			for _, m := range members {
				_, ok := m.(map[string]interface{})["status"]
				require.Equal(t, false, ok)
			}
		})
	}
}
