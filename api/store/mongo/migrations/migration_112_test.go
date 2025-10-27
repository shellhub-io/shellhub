package migrations

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
)

func TestMigration112Up(t *testing.T) {
	ctx := context.Background()

	cases := []struct {
		description string
		setup       func() error
		verify      func(tt *testing.T)
	}{
		{
			description: "succeeds removing events subdocument from sessions",
			setup: func() error {
				sessions := []bson.M{
					{
						"uid":      "session-1",
						"username": "user1",
						"events": bson.M{
							"types": []string{"shell", "pty-req"},
							"seats": []int{1, 2},
						},
					},
					{
						"uid":      "session-2",
						"username": "user2",
						"events": bson.M{
							"types": []string{"exec"},
							"seats": []int{1},
						},
					},
				}
				_, err := c.Database("test").Collection("sessions").InsertMany(ctx, []any{sessions[0], sessions[1]})

				return err
			},
			verify: func(tt *testing.T) {
				cursor, err := c.Database("test").Collection("sessions").Find(ctx, bson.M{})
				require.NoError(tt, err)

				var sessions []bson.M
				require.NoError(tt, cursor.All(ctx, &sessions))

				for _, session := range sessions {
					_, exists := session["events"]
					assert.False(tt, exists, "events field should have been removed")
				}
			},
		},
		{
			description: "succeeds when sessions have no events subdocument",
			setup: func() error {
				sessions := []bson.M{
					{
						"uid":      "session-3",
						"username": "user3",
					},
				}
				_, err := c.Database("test").Collection("sessions").InsertMany(ctx, []any{sessions[0]})

				return err
			},
			verify: func(tt *testing.T) {
				cursor, err := c.Database("test").Collection("sessions").Find(ctx, bson.M{})
				require.NoError(tt, err)

				var sessions []bson.M
				require.NoError(tt, cursor.All(ctx, &sessions))

				for _, session := range sessions {
					_, exists := session["events"]
					assert.False(tt, exists, "events field should not exist")
				}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			tt.Cleanup(func() { assert.NoError(tt, srv.Reset()) })

			require.NoError(tt, tc.setup())
			migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[111])
			require.NoError(tt, migrates.Up(ctx, migrate.AllAvailable))
			tc.verify(tt)
		})
	}
}

func TestMigration112Down(t *testing.T) {
	ctx := context.Background()

	cases := []struct {
		description string
		setup       func() error
		verify      func(tt *testing.T)
	}{
		{
			description: "succeeds reconstructing events subdocument from sessions_events",
			setup: func() error {
				sessions := []bson.M{
					{"uid": "session-1", "username": "user1"},
					{"uid": "session-2", "username": "user2"},
				}
				_, err := c.Database("test").Collection("sessions").InsertMany(ctx, []any{sessions[0], sessions[1]})
				if err != nil {
					return err
				}

				events := []bson.M{
					{"session": "session-1", "type": "shell", "seat": 1},
					{"session": "session-1", "type": "pty-req", "seat": 1},
					{"session": "session-1", "type": "shell", "seat": 2},
					{"session": "session-2", "type": "exec", "seat": 1},
				}
				_, err = c.Database("test").Collection("sessions_events").InsertMany(ctx, []any{events[0], events[1], events[2], events[3]})

				return err
			},
			verify: func(tt *testing.T) {
				migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[111])
				require.NoError(tt, migrates.Up(ctx, migrate.AllAvailable))

				require.NoError(tt, migrates.Down(ctx, migrate.AllAvailable))

				cursor, err := c.Database("test").Collection("sessions").Find(ctx, bson.M{})
				require.NoError(tt, err)

				var sessions []bson.M
				require.NoError(tt, cursor.All(ctx, &sessions))

				sessionEventsByUID := make(map[string]bson.M)
				for _, session := range sessions {
					uid := session["uid"].(string)
					sessionEventsByUID[uid] = session
				}

				session1Events := sessionEventsByUID["session-1"]["events"].(bson.M)
				types1 := session1Events["types"].(bson.A)
				seats1 := session1Events["seats"].(bson.A)

				assert.Len(tt, types1, 2)
				assert.Contains(tt, types1, "shell")
				assert.Contains(tt, types1, "pty-req")

				assert.Len(tt, seats1, 2)
				assert.Contains(tt, seats1, int32(1))
				assert.Contains(tt, seats1, int32(2))

				session2Events := sessionEventsByUID["session-2"]["events"].(bson.M)
				types2 := session2Events["types"].(bson.A)
				seats2 := session2Events["seats"].(bson.A)

				assert.Len(tt, types2, 1)
				assert.Contains(tt, types2, "exec")

				assert.Len(tt, seats2, 1)
				assert.Contains(tt, seats2, int32(1))
			},
		},
		{
			description: "succeeds creating empty events for sessions without events",
			setup: func() error {
				sessions := []bson.M{{"uid": "session-3", "username": "user3"}}
				_, err := c.Database("test").Collection("sessions").InsertMany(ctx, []any{sessions[0]})

				return err
			},
			verify: func(tt *testing.T) {
				migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[111])
				require.NoError(tt, migrates.Up(ctx, migrate.AllAvailable))

				require.NoError(tt, migrates.Down(ctx, migrate.AllAvailable))

				var session bson.M
				err := c.Database("test").Collection("sessions").FindOne(ctx, bson.M{"uid": "session-3"}).Decode(&session)
				require.NoError(tt, err)

				events := session["events"].(bson.M)
				types := events["types"].(bson.A)
				seats := events["seats"].(bson.A)

				assert.Len(tt, types, 0)
				assert.Len(tt, seats, 0)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			tt.Cleanup(func() { assert.NoError(tt, srv.Reset()) })

			require.NoError(tt, tc.setup())
			tc.verify(tt)
		})
	}
}
