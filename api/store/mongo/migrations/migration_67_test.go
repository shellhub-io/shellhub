package migrations

import (
	"context"
	"strings"
	"testing"

	"github.com/labstack/gommon/log"
	"github.com/shellhub-io/shellhub/pkg/envs"
	envMocks "github.com/shellhub-io/shellhub/pkg/envs/mocks"
	"github.com/shellhub-io/shellhub/pkg/hash"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
)

func TestMigration67Up(t *testing.T) {
	ctx := context.Background()

	mock := &envMocks.Backend{}
	envs.DefaultBackend = mock

	cases := []struct {
		description   string
		setup         func() error
		recoveryCodes []string
		test          func() error
	}{
		{
			description:   "Success to apply up on migration 67",
			recoveryCodes: []string{"secret-1", "secret-2"},
			setup: func() error {
				_, err := c.
					Database("test").
					Collection("users").
					InsertOne(ctx, models.User{
						UserData: models.UserData{
							Username: "john_doe",
						},
						MFA: models.UserMFA{
							Enabled:       true,
							RecoveryCodes: []string{"secret-1", "secret-2"},
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

			migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[66])
			require.NoError(t, migrates.Up(context.Background(), migrate.AllAvailable))

			query := c.
				Database("test").
				Collection("users").
				FindOne(context.TODO(), bson.M{"username": "john_doe"})

			user := new(models.User)
			require.NoError(t, query.Decode(user))

			log.Infof("user: %+v", user)

			require.Equal(t, len(tc.recoveryCodes), len(user.MFA.RecoveryCodes))
			for i, c := range tc.recoveryCodes {
				require.NotEqual(t, c, user.MFA.RecoveryCodes[i])
				require.Equal(t, true, strings.HasPrefix(user.MFA.RecoveryCodes[i], "$"))
				require.Equal(t, true, hash.CompareWith(c, user.MFA.RecoveryCodes[i]))
			}
		})
	}
}
