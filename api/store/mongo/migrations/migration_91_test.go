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
		verify      func(tt *testing.T)
	}{
		{
			description: "Session with single event",
			setup: func() error {
				_, err := c.
					Database("test").
					Collection("sessions").
					InsertOne(ctx, bson.M{
						"uid": "test",
						"events": bson.M{
							"types": bson.A{"test"},
							"items": []bson.M{
								{"type": "test", "data": "some data"},
							},
						},
					})

				return err
			},
			verify: func(tt *testing.T) {
				query := c.
					Database("test").
					Collection("sessions_events").
					FindOne(ctx, bson.M{"session": "test"})
				sessionEvent := make(map[string]interface{})
				require.NoError(tt, query.Decode(&sessionEvent))
				require.Contains(tt, sessionEvent, "type")
				require.Contains(tt, sessionEvent, "data")

				query = c.
					Database("test").
					Collection("sessions").
					FindOne(ctx, bson.M{"uid": "test"})
				session := make(map[string]interface{})
				require.NoError(tt, query.Decode(&session))
				require.NotContains(tt, session, "events.items")
			},
		},
		{
			description: "Session with empty events.items array",
			setup: func() error {
				_, err := c.
					Database("test").
					Collection("sessions").
					InsertOne(ctx, bson.M{
						"uid": "empty-items",
						"events": bson.M{
							"types": bson.A{"test"},
							"items": []bson.M{},
						},
					})

				return err
			},
			verify: func(tt *testing.T) {
				count, err := c.
					Database("test").
					Collection("sessions_events").
					CountDocuments(ctx, bson.M{"session": "empty-items"})
				require.NoError(tt, err)
				assert.Equal(tt, int64(0), count)

				query := c.
					Database("test").
					Collection("sessions").
					FindOne(ctx, bson.M{"uid": "empty-items"})
				session := make(map[string]interface{})
				require.NoError(tt, query.Decode(&session))
				require.NotContains(tt, session, "events.items")
			},
		},
		{
			description: "Session with multiple events",
			setup: func() error {
				_, err := c.
					Database("test").
					Collection("sessions").
					InsertOne(ctx, bson.M{
						"uid": "multi-event",
						"events": bson.M{
							"types": bson.A{"test"},
							"items": []bson.M{
								{"type": "event1", "data": "data1"},
								{"type": "event2", "data": "data2"},
							},
						},
					})

				return err
			},
			verify: func(tt *testing.T) {
				cursor, err := c.
					Database("test").
					Collection("sessions_events").
					Find(ctx, bson.M{"session": "multi-event"})
				require.NoError(tt, err)
				var events []bson.M
				require.NoError(tt, cursor.All(ctx, &events))
				assert.Equal(tt, 2, len(events))
				for _, event := range events {
					assert.Equal(tt, "multi-event", event["session"])
				}

				query := c.
					Database("test").
					Collection("sessions").
					FindOne(ctx, bson.M{"uid": "multi-event"})
				session := make(map[string]interface{})
				require.NoError(tt, query.Decode(&session))
				require.NotContains(tt, session, "events.items")
			},
		},
		{
			description: "Session with no events field",
			setup: func() error {
				_, err := c.
					Database("test").
					Collection("sessions").
					InsertOne(ctx, bson.M{
						"uid": "no-events",
					})

				return err
			},
			verify: func(tt *testing.T) {
				count, err := c.
					Database("test").
					Collection("sessions_events").
					CountDocuments(ctx, bson.M{"session": "no-events"})
				require.NoError(tt, err)
				assert.Equal(tt, int64(0), count)

				query := c.
					Database("test").
					Collection("sessions").
					FindOne(ctx, bson.M{"uid": "no-events"})
				session := make(map[string]interface{})
				require.NoError(tt, query.Decode(&session))
				_, exists := session["events"]
				assert.False(tt, exists)
			},
		},
		{
			description: "Verify indexes created",
			setup: func() error {
				_, err := c.
					Database("test").
					Collection("sessions").
					InsertOne(ctx, bson.M{
						"uid": "index-test",
						"events": bson.M{
							"types": bson.A{"test"},
							"items": []bson.M{
								{"type": "test", "data": "index data"},
							},
						},
					})

				return err
			},
			verify: func(tt *testing.T) {
				indexCursor, err := c.
					Database("test").
					Collection("sessions_events").
					Indexes().List(ctx)
				require.NoError(tt, err)
				var indexes []bson.M
				require.NoError(tt, indexCursor.All(ctx, &indexes))
				var sessionIndexFound, typeIndexFound bool
				for _, index := range indexes {
					if key, ok := index["key"].(bson.M); ok {
						if _, ok := key["session"]; ok {
							sessionIndexFound = true
						}
						if _, ok := key["type"]; ok {
							typeIndexFound = true
						}
					}
				}
				assert.True(tt, sessionIndexFound)
				assert.True(tt, typeIndexFound)
			},
		},
		{
			description: "Multiple sessions processed",
			setup: func() error {
				docs := []interface{}{
					bson.M{
						"uid": "session1",
						"events": bson.M{
							"types": bson.A{"test"},
							"items": []bson.M{
								{"type": "event1", "data": "data1"},
							},
						},
					},
					bson.M{
						"uid": "session2",
						"events": bson.M{
							"types": bson.A{"test"},
							"items": []bson.M{
								{"type": "event2", "data": "data2"},
								{"type": "event3", "data": "data3"},
							},
						},
					},
				}
				_, err := c.
					Database("test").
					Collection("sessions").
					InsertMany(ctx, docs)

				return err
			},
			verify: func(tt *testing.T) {
				count1, err := c.
					Database("test").
					Collection("sessions_events").
					CountDocuments(ctx, bson.M{"session": "session1"})
				require.NoError(tt, err)
				assert.Equal(tt, int64(1), count1)

				count2, err := c.
					Database("test").
					Collection("sessions_events").
					CountDocuments(ctx, bson.M{"session": "session2"})
				require.NoError(tt, err)
				assert.Equal(tt, int64(2), count2)

				for _, uid := range []string{"session1", "session2"} {
					query := c.
						Database("test").
						Collection("sessions").
						FindOne(ctx, bson.M{"uid": uid})
					session := make(map[string]interface{})
					require.NoError(tt, query.Decode(&session))
					require.NotContains(tt, session, "events.items")
				}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			tt.Cleanup(func() {
				assert.NoError(tt, srv.Reset())
			})
			require.NoError(tt, tc.setup())
			migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[90])
			require.NoError(tt, migrates.Up(ctx, migrate.AllAvailable))
			tc.verify(tt)
		})
	}
}
