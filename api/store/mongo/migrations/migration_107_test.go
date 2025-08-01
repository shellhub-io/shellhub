package migrations

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
)

func TestMigration107Up(t *testing.T) {
	ctx := context.Background()

	cases := []struct {
		description string
		setup       func() error
		verify      func(tt *testing.T)
	}{
		{
			description: "succeeds restructuring signon_url to binding when SAML has signon_url",
			setup: func() error {
				system := bson.M{
					"authentication": bson.M{
						"saml": bson.M{
							"enabled": true,
							"idp": bson.M{
								"signon_url": "https://example.com/saml/login",
							},
						},
					},
				}

				_, err := c.Database("test").Collection("system").InsertOne(ctx, system)

				return err
			},
			verify: func(tt *testing.T) {
				system := make(map[string]any)
				require.NoError(tt, c.Database("test").Collection("system").FindOne(ctx, bson.M{}).Decode(&system))

				auth := system["authentication"].(map[string]any)
				saml := auth["saml"].(map[string]any)
				idp := saml["idp"].(map[string]any)

				_, hasOldURL := idp["signon_url"]
				assert.False(tt, hasOldURL)

				binding, hasBinding := idp["binding"]
				require.True(tt, hasBinding)

				signonURLsMap := binding.(map[string]any)
				assert.Equal(tt, "https://example.com/saml/login", signonURLsMap["post"])
				assert.Equal(tt, "", signonURLsMap["redirect"])
				assert.Equal(tt, "post", signonURLsMap["preferred"])
			},
		},
		{
			description: "creates binding even when SAML config doesn't exist",
			setup: func() error {
				system := bson.M{
					"authentication": bson.M{
						"local": bson.M{
							"enabled": true,
						},
						"saml": bson.M{
							"enabled": false,
							"idp": bson.M{
								"signon_url": "",
							},
						},
					},
				}

				_, err := c.Database("test").Collection("system").InsertOne(ctx, system)

				return err
			},
			verify: func(tt *testing.T) {
				system := make(map[string]any)
				require.NoError(tt, c.Database("test").Collection("system").FindOne(ctx, bson.M{}).Decode(&system))

				auth := system["authentication"].(map[string]any)
				saml, hasSAML := auth["saml"]
				require.True(tt, hasSAML)

				samlMap := saml.(map[string]any)
				idp := samlMap["idp"].(map[string]any)

				binding := idp["binding"].(map[string]any)
				assert.Equal(tt, "", binding["post"])
				assert.Equal(tt, "", binding["redirect"])
				assert.Equal(tt, "", binding["preferred"])
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			tt.Cleanup(func() { assert.NoError(tt, srv.Reset()) })

			require.NoError(tt, tc.setup())
			migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[106])
			require.NoError(tt, migrates.Up(ctx, migrate.AllAvailable))
			tc.verify(tt)
		})
	}
}
