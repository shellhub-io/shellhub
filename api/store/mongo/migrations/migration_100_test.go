package migrations

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestMigration100Up(t *testing.T) {
	ctx := context.Background()
	cases := []struct {
		description string
		setup       func() error
		verify      func(tt *testing.T)
	}{
		{
			description: "remove direct-tcpip seats and types from session",
			setup: func() error {
				sessionUID := "test-session-1"
				_, err := c.Database("test").Collection("sessions").InsertOne(ctx, bson.M{
					"_id":      primitive.NewObjectID(),
					"uid":      sessionUID,
					"recorded": true,
					"events": bson.M{
						"seats": []int{0, 1, 2, 3, 4, 5},
						"types": []string{"pty-req", "direct-tcpip", "shell", "direct-tcpip", "exit-status"},
					},
				})
				if err != nil {
					return err
				}

				events := []any{
					bson.M{"_id": primitive.NewObjectID(), "session": sessionUID, "type": "pty-req", "seat": 0},
					bson.M{"_id": primitive.NewObjectID(), "session": sessionUID, "type": "direct-tcpip", "seat": 1},
					bson.M{"_id": primitive.NewObjectID(), "session": sessionUID, "type": "shell", "seat": 2},
					bson.M{"_id": primitive.NewObjectID(), "session": sessionUID, "type": "direct-tcpip", "seat": 3},
					bson.M{"_id": primitive.NewObjectID(), "session": sessionUID, "type": "exit-status", "seat": 4},
					bson.M{"_id": primitive.NewObjectID(), "session": sessionUID, "type": "direct-tcpip", "seat": 5},
				}
				_, err = c.Database("test").Collection("sessions_events").InsertMany(ctx, events)

				return err
			},
			verify: func(tt *testing.T) {
				session := make(map[string]any)
				require.NoError(tt, c.Database("test").Collection("sessions").FindOne(ctx, bson.M{"uid": "test-session-1"}).Decode(&session))

				events, ok := session["events"]
				require.True(tt, ok)
				eventsMap := events.(map[string]any)

				seats, ok := eventsMap["seats"]
				require.True(tt, ok)
				seatsList := seats.(bson.A)
				require.Len(tt, seatsList, 3)
				require.Contains(tt, seatsList, int32(0))
				require.Contains(tt, seatsList, int32(2))
				require.Contains(tt, seatsList, int32(4))
				require.NotContains(tt, seatsList, int32(1))
				require.NotContains(tt, seatsList, int32(3))
				require.NotContains(tt, seatsList, int32(5))

				types, ok := eventsMap["types"]
				require.True(tt, ok)
				typesList := types.(bson.A)
				require.Len(tt, typesList, 3)
				require.Contains(tt, typesList, "pty-req")
				require.Contains(tt, typesList, "shell")
				require.Contains(tt, typesList, "exit-status")
				require.NotContains(tt, typesList, "direct-tcpip")
			},
		},
		{
			description: "remove all direct-tcpip events from sessions_events collection",
			setup: func() error {
				events := []any{
					bson.M{"_id": primitive.NewObjectID(), "session": "session1", "type": "direct-tcpip", "seat": 1},
					bson.M{"_id": primitive.NewObjectID(), "session": "session1", "type": "shell", "seat": 2},
					bson.M{"_id": primitive.NewObjectID(), "session": "session2", "type": "direct-tcpip", "seat": 0},
				}
				_, err := c.Database("test").Collection("sessions_events").InsertMany(ctx, events)

				return err
			},
			verify: func(tt *testing.T) {
				count, err := c.Database("test").Collection("sessions_events").CountDocuments(ctx, bson.M{"type": "direct-tcpip"})
				require.NoError(tt, err)
				require.Equal(tt, int64(0), count)

				count, err = c.Database("test").Collection("sessions_events").CountDocuments(ctx, bson.M{"type": "shell"})
				require.NoError(tt, err)
				require.Equal(tt, int64(1), count)
			},
		},
		{
			description: "mark recorded=false for sessions with empty events.types after removal",
			setup: func() error {
				sessionUID := "test-session-2"
				_, err := c.Database("test").Collection("sessions").InsertOne(ctx, bson.M{
					"_id":      primitive.NewObjectID(),
					"uid":      sessionUID,
					"recorded": true,
					"events": bson.M{
						"seats": []int{0, 1},
						"types": []string{"direct-tcpip", "direct-tcpip"},
					},
				})
				if err != nil {
					return err
				}

				events := []any{
					bson.M{"_id": primitive.NewObjectID(), "session": sessionUID, "type": "direct-tcpip", "seat": 0},
					bson.M{"_id": primitive.NewObjectID(), "session": sessionUID, "type": "direct-tcpip", "seat": 1},
				}
				_, err = c.Database("test").Collection("sessions_events").InsertMany(ctx, events)

				return err
			},
			verify: func(tt *testing.T) {
				session := make(map[string]any)
				require.NoError(tt, c.Database("test").Collection("sessions").FindOne(ctx, bson.M{"uid": "test-session-2"}).Decode(&session))

				recorded, ok := session["recorded"]
				require.True(tt, ok)
				require.Equal(tt, false, recorded)

				events, ok := session["events"]
				require.True(tt, ok)
				eventsMap := events.(map[string]any)

				types, ok := eventsMap["types"]
				require.True(tt, ok)
				typesList := types.(bson.A)
				require.Len(tt, typesList, 0)

				seats, ok := eventsMap["seats"]
				require.True(tt, ok)
				seatsList := seats.(bson.A)
				require.Len(tt, seatsList, 0)
			},
		},
		{
			description: "session with no direct-tcpip events remains unchanged",
			setup: func() error {
				sessionUID := "test-session-3"
				_, err := c.Database("test").Collection("sessions").InsertOne(ctx, bson.M{
					"_id":      primitive.NewObjectID(),
					"uid":      sessionUID,
					"recorded": true,
					"events": bson.M{
						"seats": []int{0, 1, 2},
						"types": []string{"pty-req", "shell", "exit-status"},
					},
				})
				if err != nil {
					return err
				}

				events := []any{
					bson.M{"_id": primitive.NewObjectID(), "session": sessionUID, "type": "pty-req", "seat": 0},
					bson.M{"_id": primitive.NewObjectID(), "session": sessionUID, "type": "shell", "seat": 1},
					bson.M{"_id": primitive.NewObjectID(), "session": sessionUID, "type": "exit-status", "seat": 2},
				}
				_, err = c.Database("test").Collection("sessions_events").InsertMany(ctx, events)

				return err
			},
			verify: func(tt *testing.T) {
				session := make(map[string]any)
				require.NoError(tt, c.Database("test").Collection("sessions").FindOne(ctx, bson.M{"uid": "test-session-3"}).Decode(&session))

				recorded, ok := session["recorded"]
				require.True(tt, ok)
				require.Equal(tt, true, recorded)

				events, ok := session["events"]
				require.True(tt, ok)
				eventsMap := events.(map[string]any)

				seats, ok := eventsMap["seats"]
				require.True(tt, ok)
				seatsList := seats.(bson.A)
				require.Len(tt, seatsList, 3)
				require.Contains(tt, seatsList, int32(0))
				require.Contains(tt, seatsList, int32(1))
				require.Contains(tt, seatsList, int32(2))

				types, ok := eventsMap["types"]
				require.True(tt, ok)
				typesList := types.(bson.A)
				require.Len(tt, typesList, 3)
				require.Contains(tt, typesList, "pty-req")
				require.Contains(tt, typesList, "shell")
				require.Contains(tt, typesList, "exit-status")
			},
		},
		{
			description: "partial removal - some seats have direct-tcpip, others don't",
			setup: func() error {
				sessionUID := "test-session-4"
				_, err := c.Database("test").Collection("sessions").InsertOne(ctx, bson.M{
					"_id":      primitive.NewObjectID(),
					"uid":      sessionUID,
					"recorded": true,
					"events": bson.M{
						"seats": []int{0, 1, 2, 3, 4},
						"types": []string{"pty-req", "direct-tcpip", "shell", "exit-status"},
					},
				})
				if err != nil {
					return err
				}

				events := []any{
					bson.M{"_id": primitive.NewObjectID(), "session": sessionUID, "type": "pty-req", "seat": 0},
					bson.M{"_id": primitive.NewObjectID(), "session": sessionUID, "type": "direct-tcpip", "seat": 1},
					bson.M{"_id": primitive.NewObjectID(), "session": sessionUID, "type": "shell", "seat": 2},
					bson.M{"_id": primitive.NewObjectID(), "session": sessionUID, "type": "exit-status", "seat": 3},
					bson.M{"_id": primitive.NewObjectID(), "session": sessionUID, "type": "direct-tcpip", "seat": 4},
				}
				_, err = c.Database("test").Collection("sessions_events").InsertMany(ctx, events)

				return err
			},
			verify: func(tt *testing.T) {
				session := make(map[string]any)
				require.NoError(tt, c.Database("test").Collection("sessions").FindOne(ctx, bson.M{"uid": "test-session-4"}).Decode(&session))

				recorded, ok := session["recorded"]
				require.True(tt, ok)
				require.Equal(tt, true, recorded)

				events, ok := session["events"]
				require.True(tt, ok)
				eventsMap := events.(map[string]any)

				seats, ok := eventsMap["seats"]
				require.True(tt, ok)
				seatsList := seats.(bson.A)
				require.Len(tt, seatsList, 3)
				require.Contains(tt, seatsList, int32(0))
				require.Contains(tt, seatsList, int32(2))
				require.Contains(tt, seatsList, int32(3))
				require.NotContains(tt, seatsList, int32(1))
				require.NotContains(tt, seatsList, int32(4))

				types, ok := eventsMap["types"]
				require.True(tt, ok)
				typesList := types.(bson.A)
				require.Len(tt, typesList, 3)
				require.Contains(tt, typesList, "pty-req")
				require.Contains(tt, typesList, "shell")
				require.Contains(tt, typesList, "exit-status")
				require.NotContains(tt, typesList, "direct-tcpip")
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			tt.Cleanup(func() {
				assert.NoError(tt, srv.Reset())
			})
			require.NoError(tt, tc.setup())
			migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[99])
			require.NoError(tt, migrates.Up(ctx, migrate.AllAvailable))
			tc.verify(tt)
		})
	}
}

func TestMigration100Down(t *testing.T) {
	ctx := context.Background()
	cases := []struct {
		description string
		setup       func() error
		verify      func(tt *testing.T)
	}{
		{
			description: "revert recorded=true for sessions with empty events.types",
			setup: func() error {
				sessionUID := "test-session-1"
				_, err := c.Database("test").Collection("sessions").InsertOne(ctx, bson.M{
					"_id":      primitive.NewObjectID(),
					"uid":      sessionUID,
					"recorded": true,
					"events": bson.M{
						"seats": []int{0, 1},
						"types": []string{"direct-tcpip"},
					},
				})
				if err != nil {
					return err
				}

				events := []any{
					bson.M{"_id": primitive.NewObjectID(), "session": sessionUID, "type": "direct-tcpip", "seat": 0},
					bson.M{"_id": primitive.NewObjectID(), "session": sessionUID, "type": "direct-tcpip", "seat": 1},
				}
				_, err = c.Database("test").Collection("sessions_events").InsertMany(ctx, events)

				return err
			},
			verify: func(tt *testing.T) {
				session := make(map[string]any)
				require.NoError(tt, c.Database("test").Collection("sessions").FindOne(ctx, bson.M{"uid": "test-session-1"}).Decode(&session))

				recorded, ok := session["recorded"]
				require.True(tt, ok)
				require.Equal(tt, true, recorded)

				events, ok := session["events"]
				require.True(tt, ok)
				eventsMap := events.(map[string]any)

				types, ok := eventsMap["types"]
				require.True(tt, ok)
				typesList := types.(bson.A)
				require.Len(tt, typesList, 0)

				seats, ok := eventsMap["seats"]
				require.True(tt, ok)
				seatsList := seats.(bson.A)
				require.Len(tt, seatsList, 0)
			},
		},
		{
			description: "verify direct-tcpip events cannot be restored in sessions_events",
			setup: func() error {
				events := []any{
					bson.M{"_id": primitive.NewObjectID(), "session": "session1", "type": "direct-tcpip", "seat": 1},
					bson.M{"_id": primitive.NewObjectID(), "session": "session1", "type": "shell", "seat": 2},
				}
				_, err := c.Database("test").Collection("sessions_events").InsertMany(ctx, events)

				return err
			},
			verify: func(tt *testing.T) {
				count, err := c.Database("test").Collection("sessions_events").CountDocuments(ctx, bson.M{"type": "direct-tcpip"})
				require.NoError(tt, err)
				require.Equal(tt, int64(0), count)

				count, err = c.Database("test").Collection("sessions_events").CountDocuments(ctx, bson.M{"type": "shell"})
				require.NoError(tt, err)
				require.Equal(tt, int64(1), count)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			tt.Cleanup(func() {
				assert.NoError(tt, srv.Reset())
			})
			require.NoError(tt, tc.setup())
			migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[99])
			require.NoError(tt, migrates.Up(ctx, migrate.AllAvailable))
			require.NoError(tt, migrates.Down(ctx, migrate.AllAvailable))
			tc.verify(tt)
		})
	}
}
