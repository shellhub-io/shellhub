package migrations

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/envs"
	envMocks "github.com/shellhub-io/shellhub/pkg/envs/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestMigration66Up(t *testing.T) {
	ctx := context.Background()

	mock := &envMocks.Backend{}
	envs.DefaultBackend = mock

	cases := []struct {
		description string
		setup       func() error
		test        func() error
		expected    map[string]interface{}
	}{
		{
			description: "Success to apply up on migration 66",
			setup: func() error {
				_, err := c.
					Database("test").
					Collection("users").
					InsertOne(ctx, map[string]interface{}{
						"username":   "john_doe",
						"status_mfa": true,
						"secret":     "secret",
						"codes":      []string{"code-1", "code-2"},
					})

				return err
			},
			expected: map[string]interface{}{
				"mfa": map[string]interface{}{
					"enabled":        true,
					"secret":         "secret",
					"recovery_codes": primitive.A{"code-1", "code-2"},
				},
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

			migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[65])
			require.NoError(t, migrates.Up(context.Background(), migrate.AllAvailable))

			query := c.
				Database("test").
				Collection("users").
				FindOne(context.TODO(), bson.M{"username": "john_doe"})

			user := make(map[string]interface{})
			require.NoError(t, query.Decode(&user))

			attr, ok := user["mfa"]
			require.Equal(t, true, ok)
			require.Equal(t, tc.expected["mfa"], attr)

			_, ok = user["status_mfa"]
			require.Equal(t, false, ok)
			_, ok = user["secret"]
			require.Equal(t, false, ok)
			_, ok = user["codes"]
			require.Equal(t, false, ok)
		})
	}
}
