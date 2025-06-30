package migrations

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
)

func TestMigration99Up(t *testing.T) {
	ctx := context.Background()

	cases := []struct {
		description string
		setup       func() error
		verify      func(tt *testing.T)
	}{
		{
			description: "mark recorded to false for sessions without event types",
			setup: func() error {
				if _, err := c.
					Database("test").
					Collection("sessions").
					InsertOne(ctx, bson.M{
						"uid":      "first",
						"recorded": true,
						"events":   bson.M{"types": []any{}},
					}); err != nil {
					return err
				}

				if _, err := c.
					Database("test").
					Collection("sessions").
					InsertOne(ctx, bson.M{
						"uid":      "second",
						"recorded": true,
						"events": bson.M{"types": []any{
							"shell",
						}},
					}); err != nil {
					return err
				}

				return nil
			},
			verify: func(tt *testing.T) {
				sessions := []map[string]any{}

				cursor, err := c.Database("test").Collection("sessions").Find(ctx, bson.M{})
				require.NoError(tt, err)

				ctx := context.Background()

				err = cursor.All(ctx, &sessions)
				require.NoError(tt, err)

				recorded, ok := sessions[0]["recorded"]
				require.True(tt, ok)
				require.Equal(tt, false, recorded)

				recorded, ok = sessions[1]["recorded"]
				require.True(tt, ok)
				require.Equal(tt, true, recorded)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			tt.Cleanup(func() {
				assert.NoError(tt, srv.Reset())
			})

			require.NoError(tt, tc.setup())

			migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[98])
			require.NoError(tt, migrates.Up(ctx, migrate.AllAvailable))

			tc.verify(tt)
		})
	}
}
