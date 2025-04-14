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

func TestMigration92Up(t *testing.T) {
	ctx := context.Background()
	mock := &envmock.Backend{}
	envs.DefaultBackend = mock

	cases := []struct {
		description string
		setup       func() error
		verify      func(tt *testing.T)
	}{
		{
			description: "Authenticated session should get events.seats set to [int32(0)]",
			setup: func() error {
				_, err := c.
					Database("test").
					Collection("sessions").
					InsertOne(ctx, bson.M{
						"uid":           "auth-test",
						"authenticated": true,
						"events": bson.M{
							"types": bson.A{"test"},
						},
					})
				if err != nil {
					return err
				}
				_, err = c.
					Database("test").
					Collection("sessions_events").
					InsertOne(ctx, bson.M{
						"session": "auth-test",
						"type":    "test",
						"data":    "some data",
					})

				return err
			},
			verify: func(tt *testing.T) {
				query := c.
					Database("test").
					Collection("sessions").
					FindOne(ctx, bson.M{"uid": "auth-test"})
				session := make(map[string]interface{})
				require.NoError(tt, query.Decode(&session))
				events, ok := session["events"].(map[string]interface{})
				require.True(tt, ok)
				seats, exists := events["seats"]
				require.True(tt, exists)
				assert.Equal(tt, bson.A{int32(0)}, seats)

				query = c.
					Database("test").
					Collection("sessions_events").
					FindOne(ctx, bson.M{"session": "auth-test"})
				sessionEvent := make(map[string]interface{})
				require.NoError(tt, query.Decode(&sessionEvent))
				seat, exists := sessionEvent["seat"]
				require.True(tt, exists)
				assert.Equal(tt, int32(0), seat)
			},
		},
		{
			description: "Unauthenticated session should get events.seats set to []",
			setup: func() error {
				_, err := c.
					Database("test").
					Collection("sessions").
					InsertOne(ctx, bson.M{
						"uid":           "unauth-test",
						"authenticated": false,
						"events": bson.M{
							"types": bson.A{"test"},
						},
					})
				if err != nil {
					return err
				}
				_, err = c.
					Database("test").
					Collection("sessions_events").
					InsertOne(ctx, bson.M{
						"session": "unauth-test",
						"type":    "test",
						"data":    "some data",
					})

				return err
			},
			verify: func(tt *testing.T) {
				query := c.
					Database("test").
					Collection("sessions").
					FindOne(ctx, bson.M{"uid": "unauth-test"})
				session := make(map[string]interface{})
				require.NoError(tt, query.Decode(&session))
				events, ok := session["events"].(map[string]interface{})
				require.True(tt, ok)
				seats, exists := events["seats"]
				require.True(tt, exists)
				assert.Equal(tt, bson.A{}, seats)

				query = c.
					Database("test").
					Collection("sessions_events").
					FindOne(ctx, bson.M{"session": "unauth-test"})
				sessionEvent := make(map[string]interface{})
				require.NoError(tt, query.Decode(&sessionEvent))
				seat, exists := sessionEvent["seat"]
				require.True(tt, exists)
				assert.Equal(tt, int32(0), seat)
			},
		},
		{
			description: "Session without events field should be updated if authenticated",
			setup: func() error {
				_, err := c.
					Database("test").
					Collection("sessions").
					InsertOne(ctx, bson.M{
						"uid":           "no-events-test",
						"authenticated": true,
					})
				if err != nil {
					return err
				}
				_, err = c.
					Database("test").
					Collection("sessions_events").
					InsertOne(ctx, bson.M{
						"session": "no-events-test",
						"type":    "test",
						"data":    "some data",
					})

				return err
			},
			verify: func(tt *testing.T) {
				query := c.
					Database("test").
					Collection("sessions").
					FindOne(ctx, bson.M{"uid": "no-events-test"})
				session := make(map[string]interface{})
				require.NoError(tt, query.Decode(&session))
				events, ok := session["events"].(map[string]interface{})
				require.True(tt, ok)
				seats, exists := events["seats"]
				require.True(tt, exists)
				assert.Equal(tt, bson.A{int32(0)}, seats)

				query = c.
					Database("test").
					Collection("sessions_events").
					FindOne(ctx, bson.M{"session": "no-events-test"})
				sessionEvent := make(map[string]interface{})
				require.NoError(tt, query.Decode(&sessionEvent))
				seat, exists := sessionEvent["seat"]
				require.True(tt, exists)
				assert.Equal(tt, int32(0), seat)
			},
		},
		{
			description: "Multiple sessions documents update correctly",
			setup: func() error {
				_, err := c.
					Database("test").
					Collection("sessions").
					InsertMany(ctx, []interface{}{
						bson.M{
							"uid":           "multi-test-auth",
							"authenticated": true,
							"events":        bson.M{"types": bson.A{"test"}},
						},
						bson.M{
							"uid":           "multi-test-unauth",
							"authenticated": false,
							"events":        bson.M{"types": bson.A{"test"}},
						},
					})
				if err != nil {
					return err
				}
				_, err = c.
					Database("test").
					Collection("sessions_events").
					InsertMany(ctx, []interface{}{
						bson.M{
							"session": "multi-test-auth",
							"type":    "test",
							"data":    "data1",
						},
						bson.M{
							"session": "multi-test-unauth",
							"type":    "test",
							"data":    "data2",
						},
					})

				return err
			},
			verify: func(tt *testing.T) {
				query := c.
					Database("test").
					Collection("sessions").
					FindOne(ctx, bson.M{"uid": "multi-test-auth"})
				sessionAuth := make(map[string]interface{})
				require.NoError(tt, query.Decode(&sessionAuth))
				eventsAuth, ok := sessionAuth["events"].(map[string]interface{})
				require.True(tt, ok)
				seatsAuth, exists := eventsAuth["seats"]
				require.True(tt, exists)
				assert.Equal(tt, bson.A{int32(0)}, seatsAuth)

				query = c.
					Database("test").
					Collection("sessions").
					FindOne(ctx, bson.M{"uid": "multi-test-unauth"})
				sessionUnauth := make(map[string]interface{})
				require.NoError(tt, query.Decode(&sessionUnauth))
				eventsUnauth, ok := sessionUnauth["events"].(map[string]interface{})
				require.True(tt, ok)
				seatsUnauth, exists := eventsUnauth["seats"]
				require.True(tt, exists)
				assert.Equal(tt, bson.A{}, seatsUnauth)

				query = c.
					Database("test").
					Collection("sessions_events").
					FindOne(ctx, bson.M{"session": "multi-test-auth"})
				sessionEventAuth := make(map[string]interface{})
				require.NoError(tt, query.Decode(&sessionEventAuth))
				seatAuth, exists := sessionEventAuth["seat"]
				require.True(tt, exists)
				assert.Equal(tt, int32(0), seatAuth)

				query = c.
					Database("test").
					Collection("sessions_events").
					FindOne(ctx, bson.M{"session": "multi-test-unauth"})
				sessionEventUnauth := make(map[string]interface{})
				require.NoError(tt, query.Decode(&sessionEventUnauth))
				seatUnauth, exists := sessionEventUnauth["seat"]
				require.True(tt, exists)
				assert.Equal(tt, int32(0), seatUnauth)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			tt.Cleanup(func() {
				assert.NoError(tt, srv.Reset())
			})

			require.NoError(tt, tc.setup())

			migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[91])
			require.NoError(tt, migrates.Up(ctx, migrate.AllAvailable))

			tc.verify(tt)
		})
	}
}
