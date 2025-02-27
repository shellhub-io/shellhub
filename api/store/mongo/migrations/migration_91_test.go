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

func TestMigration91Up(t *testing.T) {
	ctx := context.Background()
	mock := &envmock.Backend{}
	envs.DefaultBackend = mock

	cases := []struct {
		description string
		setup       func() error
	}{
		{
			description: "Success to apply up on migration 91",
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
						},
					})
				if err != nil {
					return err
				}

				_, err = c.
					Database("test").
					Collection("sessions_events").
					InsertOne(ctx, bson.M{
						"session": "session-1",
						"type":    "test",
						"data":    "some data",
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

			migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[90])
			require.NoError(tt, migrates.Up(ctx, migrate.AllAvailable))

			query := c.
				Database("test").
				Collection("sessions").
				FindOne(ctx, bson.M{"uid": "session-1"})
			session := make(map[string]interface{})
			require.NoError(tt, query.Decode(&session))

			events, ok := session["events"].(map[string]interface{})
			require.True(tt, ok, "events field should exist")

			seats, ok := events["seats"].(bson.A)
			require.True(tt, ok, "events.seats field should exist")
			require.Equal(tt, 1, len(seats), "seats array should have one element")
			require.Equal(tt, int32(0), seats[0], "first seat should be 0")

			query = c.
				Database("test").
				Collection("sessions_events").
				FindOne(ctx, bson.M{"session": "session-1"})
			sessionEvent := make(map[string]interface{})
			require.NoError(tt, query.Decode(&sessionEvent))

			seat, ok := sessionEvent["seat"]
			require.True(tt, ok, "seat field should exist")
			require.Equal(tt, int32(0), seat, "seat should be 0")
		})
	}
}
