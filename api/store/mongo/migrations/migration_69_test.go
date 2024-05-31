package migrations

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/envs"
	envMocks "github.com/shellhub-io/shellhub/pkg/envs/mocks"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	mongodb "go.mongodb.org/mongo-driver/mongo"
)

func TestMigration69Up(t *testing.T) {
	ctx := context.Background()

	mock := &envMocks.Backend{}
	envs.DefaultBackend = mock

	cases := []struct {
		description string
		setup       func() error
		plainID     string
		test        func() error
	}{
		{
			description: "Success to apply up on migration 69",
			plainID:     "343d67d3-5084-4845-ab10-59891c88ec76",
			setup: func() error {
				_, err := c.
					Database("test").
					Collection("api_keys").
					InsertOne(ctx, models.APIKey{ID: "343d67d3-5084-4845-ab10-59891c88ec76"})

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

			migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[68])
			require.NoError(t, migrates.Up(context.Background(), migrate.AllAvailable))

			idSum := sha256.Sum256([]byte(tc.plainID))
			hashedID := hex.EncodeToString(idSum[:])

			old := c.
				Database("test").
				Collection("api_keys").
				FindOne(context.TODO(), bson.M{"_id": tc.plainID}).
				Decode(&models.APIKey{})
			require.Equal(t, mongodb.ErrNoDocuments, old)

			query := c.
				Database("test").
				Collection("api_keys").
				FindOne(context.TODO(), bson.M{"_id": hashedID})

			apiKey := new(models.APIKey)
			require.NoError(t, query.Decode(apiKey))
			require.Equal(t, hashedID, apiKey.ID)
		})
	}
}
