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

func TestMigration95Up(t *testing.T) {
	ctx := context.Background()
	mock := &envmock.Backend{}
	envs.DefaultBackend = mock

	mock.On("Get", "SHELLHUB_ENTERPRISE").Return("true").Once()

	cases := []struct {
		description string
		setup       func() error
		verify      func(tt *testing.T)
	}{
		{
			description: "Single recorded session should be converted to events",
			setup: func() error {
				_, err := c.
					Database("test").
					Collection("sessions").
					InsertOne(ctx, bson.M{
						"uid":           "session-1",
						"authenticated": true,
						"events": bson.M{
							"types": bson.A{},
							"seats": bson.A{},
						},
					})
				if err != nil {
					return err
				}

				_, err = c.
					Database("test").
					Collection("recorded_sessions").
					InsertMany(ctx, []any{
						bson.M{
							"uid":     "session-1",
							"message": "initial output",
							"time":    "2023-01-01T10:00:00Z",
							"width":   80,
							"height":  24,
						},
						bson.M{
							"uid":     "session-1",
							"message": "resized terminal output",
							"time":    "2023-01-01T10:01:00Z",
							"width":   100,
							"height":  30,
						},
						bson.M{
							"uid":     "session-1",
							"message": "final output",
							"time":    "2023-01-01T10:02:00Z",
							"width":   100,
							"height":  30,
						},
					})

				return err
			},
			verify: func(tt *testing.T) {
				query := c.
					Database("test").
					Collection("sessions").
					FindOne(ctx, bson.M{"uid": "session-1"})

				session := make(map[string]any)
				require.NoError(tt, query.Decode(&session))

				events, ok := session["events"].(map[string]any)
				require.True(tt, ok)

				types, ok := events["types"].(bson.A)
				require.True(tt, ok)
				assert.Contains(tt, types, "pty-req")
				assert.Contains(tt, types, "window-change")
				assert.Contains(tt, types, "pty-output")

				seats, ok := events["seats"].(bson.A)
				require.True(tt, ok)
				assert.Contains(tt, seats, int32(0))

				cursor, err := c.
					Database("test").
					Collection("sessions_events").
					Find(ctx, bson.M{"session": "session-1"})
				require.NoError(tt, err)
				defer cursor.Close(ctx)

				var sessionEvents []map[string]any
				require.NoError(tt, cursor.All(ctx, &sessionEvents))

				assert.Equal(tt, 5, len(sessionEvents))

				eventTypeCounts := make(map[string]int)
				for _, event := range sessionEvents {
					eventType := event["type"].(string)
					eventTypeCounts[eventType]++
					assert.Equal(tt, int32(0), event["seat"])
				}

				assert.Equal(tt, 1, eventTypeCounts["pty-req"])
				assert.Equal(tt, 1, eventTypeCounts["window-change"])
				assert.Equal(tt, 3, eventTypeCounts["pty-output"])

				count, err := c.
					Database("test").
					Collection("recorded_sessions").
					CountDocuments(ctx, bson.M{"uid": "session-1"})
				require.NoError(tt, err)
				assert.Equal(tt, int64(0), count)
			},
		},
		{
			description: "When not in enterprise mode, nothing should be migrated",
			setup: func() error {
				mock.On("Get", "SHELLHUB_ENTERPRISE").Return("false").Once()

				_, err := c.
					Database("test").
					Collection("sessions").
					InsertOne(ctx, bson.M{
						"uid":           "session-4",
						"authenticated": true,
						"events": bson.M{
							"types": bson.A{},
							"seats": bson.A{},
						},
					})
				if err != nil {
					return err
				}

				_, err = c.
					Database("test").
					Collection("recorded_sessions").
					InsertOne(ctx, bson.M{
						"uid":     "session-4",
						"message": "test output",
						"time":    "2023-01-01T10:00:00Z",
						"width":   80,
						"height":  24,
					})

				return err
			},
			verify: func(tt *testing.T) {
				count, err := c.
					Database("test").
					Collection("recorded_sessions").
					CountDocuments(ctx, bson.M{"uid": "session-4"})
				require.NoError(tt, err)
				assert.Equal(tt, int64(1), count)

				count, err = c.
					Database("test").
					Collection("sessions_events").
					CountDocuments(ctx, bson.M{"session": "session-4"})
				require.NoError(tt, err)
				assert.Equal(tt, int64(0), count)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			tt.Cleanup(func() {
				assert.NoError(tt, srv.Reset())
			})

			require.NoError(tt, tc.setup())

			migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[94])
			require.NoError(tt, migrates.Up(ctx, migrate.AllAvailable))

			tc.verify(tt)
		})
	}
}

func TestMigration95Down(t *testing.T) {
	ctx := context.Background()
	mock := &envmock.Backend{}
	envs.DefaultBackend = mock

	mock.On("Get", "SHELLHUB_ENTERPRISE").Return("true").Twice()

	cases := []struct {
		description string
		setup       func() error
		verify      func(tt *testing.T)
	}{
		{
			description: "Success to revert migration 95 in cloud mode",
			setup: func() error {
				_, err := c.
					Database("test").
					Collection("sessions").
					InsertOne(ctx, bson.M{
						"uid":           "session-1",
						"authenticated": true,
						"events": bson.M{
							"types": bson.A{"pty-req", "window-change", "pty-output"},
							"seats": bson.A{int32(0)},
						},
					})
				if err != nil {
					return err
				}

				timestamp1 := "2023-01-01T10:00:00Z"
				timestamp2 := "2023-01-01T10:01:00Z"
				timestamp3 := "2023-01-01T10:02:00Z"

				_, err = c.
					Database("test").
					Collection("sessions_events").
					InsertMany(ctx, []any{
						bson.M{
							"session":   "session-1",
							"type":      "pty-req",
							"timestamp": timestamp1,
							"data": bson.M{
								"term":     "",
								"columns":  uint32(80),
								"rows":     uint32(24),
								"width":    uint32(0),
								"height":   uint32(0),
								"modelist": []byte{},
							},
							"seat": int32(0),
						},
						bson.M{
							"session":   "session-1",
							"type":      "window-change",
							"timestamp": timestamp2,
							"data": bson.M{
								"columns": uint32(100),
								"rows":    uint32(30),
								"width":   uint32(0),
								"height":  uint32(0),
							},
							"seat": int32(0),
						},
						bson.M{
							"session":   "session-1",
							"type":      "pty-output",
							"timestamp": timestamp1,
							"data": bson.M{
								"output": "initial output",
							},
							"seat": int32(0),
						},
						bson.M{
							"session":   "session-1",
							"type":      "pty-output",
							"timestamp": timestamp2,
							"data": bson.M{
								"output": "resized terminal output",
							},
							"seat": int32(0),
						},
						bson.M{
							"session":   "session-1",
							"type":      "pty-output",
							"timestamp": timestamp3,
							"data": bson.M{
								"output": "final output",
							},
							"seat": int32(0),
						},
					})

				return err
			},
			verify: func(tt *testing.T) {
				query := c.
					Database("test").
					Collection("sessions").
					FindOne(ctx, bson.M{"uid": "session-1"})

				session := make(map[string]any)
				require.NoError(tt, query.Decode(&session))

				events, ok := session["events"].(map[string]any)
				require.True(tt, ok)

				types, ok := events["types"].(bson.A)
				require.True(tt, ok)
				assert.NotContains(tt, types, "pty-req")
				assert.NotContains(tt, types, "window-change")
				assert.NotContains(tt, types, "pty-output")

				count, err := c.
					Database("test").
					Collection("sessions_events").
					CountDocuments(ctx, bson.M{
						"session": "session-1",
						"type": bson.M{
							"$in": []string{"pty-req", "window-change", "pty-output"},
						},
					})
				require.NoError(tt, err)
				assert.Equal(tt, int64(0), count)

				cursor, err := c.
					Database("test").
					Collection("recorded_sessions").
					Find(ctx, bson.M{"uid": "session-1"})
				require.NoError(tt, err)
				defer cursor.Close(ctx)

				var recordedSessions []map[string]any
				require.NoError(tt, cursor.All(ctx, &recordedSessions))

				assert.Equal(tt, 3, len(recordedSessions))

				widthHeightCount := make(map[string]int)
				for _, record := range recordedSessions {
					key := fmt.Sprintf("%v-%v", record["width"], record["height"])
					widthHeightCount[key]++
					assert.Contains(tt, []string{
						"initial output",
						"resized terminal output",
						"final output",
					}, record["message"])
				}

				assert.Equal(tt, 1, widthHeightCount["80-24"])
				assert.Equal(tt, 2, widthHeightCount["100-30"])
			},
		},
		{
			description: "Skip migration revert when not in enterprise mode",
			setup: func() error {
				mock.On("Get", "SHELLHUB_ENTERPRISE").Return("false").Twice()

				_, err := c.
					Database("test").
					Collection("sessions").
					InsertOne(ctx, bson.M{
						"uid":           "session-3",
						"authenticated": true,
						"events": bson.M{
							"types": bson.A{"pty-req", "pty-output"},
							"seats": bson.A{int32(0)},
						},
					})
				if err != nil {
					return err
				}

				_, err = c.
					Database("test").
					Collection("sessions_events").
					InsertOne(ctx, bson.M{
						"session":   "session-3",
						"type":      "pty-output",
						"timestamp": "2023-01-01T10:00:00Z",
						"data": bson.M{
							"output": "test output",
						},
						"seat": int32(0),
					})

				return err
			},
			verify: func(tt *testing.T) {
				query := c.
					Database("test").
					Collection("sessions").
					FindOne(ctx, bson.M{"uid": "session-3"})

				session := make(map[string]any)
				require.NoError(tt, query.Decode(&session))

				events, ok := session["events"].(map[string]any)
				require.True(tt, ok)

				types, ok := events["types"].(bson.A)
				require.True(tt, ok)
				assert.Contains(tt, types, "pty-req")
				assert.Contains(tt, types, "pty-output")

				count, err := c.
					Database("test").
					Collection("sessions_events").
					CountDocuments(ctx, bson.M{
						"session": "session-3",
						"type":    "pty-output",
					})
				require.NoError(tt, err)
				assert.Equal(tt, int64(1), count)

				count, err = c.
					Database("test").
					Collection("recorded_sessions").
					CountDocuments(ctx, bson.M{"uid": "session-3"})
				require.NoError(tt, err)
				assert.Equal(tt, int64(0), count)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			tt.Cleanup(func() {
				assert.NoError(tt, srv.Reset())
			})

			require.NoError(tt, tc.setup())

			migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[94])
			require.NoError(tt, migrates.Up(ctx, migrate.AllAvailable))

			require.NoError(tt, migrates.Down(ctx, migrate.AllAvailable))

			tc.verify(tt)
		})
	}
}
