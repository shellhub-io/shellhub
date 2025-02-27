package migrations

import (
	"context"
	"fmt"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/envs"
	envmock "github.com/shellhub-io/shellhub/pkg/envs/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
)

func TestMigration90Up(t *testing.T) {
	ctx := context.Background()

	mock := &envmock.Backend{}
	envs.DefaultBackend = mock

	cases := []struct {
		description string
		setup       func() error
	}{
		{
			description: "Success to apply up on migration 90",
			setup: func() error {
				_, err := c.
					Database("test").
					Collection("sessions").
					InsertOne(ctx, bson.M{
						"uid": "session-1",
						"events": bson.M{
							"types": bson.A{
								"test",
							},
							"items": []bson.M{
								{
									"type": "test",
									"data": "some data",
								},
							},
						},
					})

				return err
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			tt.Cleanup(func() {
				assert.NoError(tt, srv.Reset())
			})

			assert.NoError(tt, tc.setup())

			migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[89])
			require.NoError(tt, migrates.Up(ctx, migrate.AllAvailable))

			query := c.
				Database("test").
				Collection("sessions_events").
				FindOne(ctx, bson.M{"session": "session-1"})

			sessionEvent := make(map[string]interface{})
			require.NoError(tt, query.Decode(&sessionEvent))
			fmt.Println(sessionEvent)

			require.Contains(tt, sessionEvent, "type")
			require.Contains(tt, sessionEvent, "data")

			query = c.
				Database("test").
				Collection("sessions").
				FindOne(ctx, bson.M{"uid": "session-1"})

			session := make(map[string]interface{})
			require.NoError(tt, query.Decode(&session))
			fmt.Println(session)

			require.NotContains(tt, session, "events.items")
		})
	}
}
