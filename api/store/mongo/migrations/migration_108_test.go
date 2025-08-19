package migrations

import (
	"context"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
)

func TestMigration108Up(t *testing.T) {
	ctx := context.Background()

	cases := []struct {
		description string
		setup       func() error
		verify      func(t *testing.T)
	}{
		{
			description: "succeeds migrating session events to seats",
			setup: func() error {
				sessions := []bson.M{
					{
						"uid": "session-1",
						"events": bson.M{
							"types": []string{"pty-output", "window-change"},
							"seats": []int{0, 1},
						},
					},
					{
						"uid": "session-2",
						"events": bson.M{
							"types": []string{"pty-output"},
							"seats": []int{0},
						},
					},
				}

				events := []bson.M{
					{
						"session":   "session-1",
						"type":      "pty-output",
						"seat":      0,
						"timestamp": time.Now(),
					},
					{
						"session":   "session-1",
						"type":      "window-change",
						"seat":      0,
						"timestamp": time.Now(),
					},
					{
						"session":   "session-1",
						"type":      "pty-output",
						"seat":      1,
						"timestamp": time.Now(),
					},
					{
						"session":   "session-2",
						"type":      "pty-output",
						"seat":      0,
						"timestamp": time.Now(),
					},
				}

				if _, err := c.Database("test").Collection("sessions").InsertMany(ctx, []any{sessions[0], sessions[1]}); err != nil {
					return err
				}

				if _, err := c.Database("test").Collection("sessions_events").InsertMany(ctx, []any{events[0], events[1], events[2], events[3]}); err != nil {
					return err
				}

				return nil
			},
			verify: func(t *testing.T) {
				var session1 struct {
					Events models.SessionEvents `bson:"events"`
					Seats  []models.SessionSeat `bson:"seats"`
				}

				err := c.Database("test").Collection("sessions").FindOne(ctx, bson.M{"uid": "session-1"}).Decode(&session1)
				assert.NoError(t, err)

				assert.ElementsMatch(t, []string{"pty-output", "window-change"}, session1.Events.Types)
				assert.ElementsMatch(t, []int{0, 1}, session1.Events.Seats)

				assert.Len(t, session1.Seats, 2)
				for _, seat := range session1.Seats {
					switch seat.ID {
					case 0:
						assert.ElementsMatch(t, []string{"pty-output", "window-change"}, seat.Events)
					case 1:
						assert.ElementsMatch(t, []string{"pty-output"}, seat.Events)
					}
				}

				var session2 struct {
					Events models.SessionEvents `bson:"events"`
					Seats  []models.SessionSeat `bson:"seats"`
				}

				err = c.Database("test").Collection("sessions").FindOne(ctx, bson.M{"uid": "session-2"}).Decode(&session2)
				assert.NoError(t, err)

				assert.ElementsMatch(t, []string{"pty-output"}, session2.Events.Types)
				assert.ElementsMatch(t, []int{0}, session2.Events.Seats)

				assert.Len(t, session2.Seats, 1)
				for _, seat := range session2.Seats {
					switch seat.ID {
					case 0:
						assert.ElementsMatch(t, []string{"pty-output"}, seat.Events)
					}
				}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			t.Cleanup(func() { assert.NoError(t, srv.Reset()) })

			require.NoError(t, tc.setup())
			migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[107])
			require.NoError(t, migrates.Up(ctx, migrate.AllAvailable))

			tc.verify(t)
		})
	}
}

func TestMigration108Down(t *testing.T) {
	ctx := context.Background()

	cases := []struct {
		description string
		setup       func() error
		verify      func(t *testing.T)
	}{
		{
			description: "succeeds removing events field while keeping seats structure",
			setup: func() error {
				sessions := []bson.M{
					{
						"uid": "session-1",
						"events": bson.M{
							"types": []string{"pty-output", "window-change"},
							"seats": []int{0, 1},
						},
						"seats": []bson.M{
							{
								"id":     0,
								"events": []string{"pty-output", "window-change"},
							},
							{
								"id":     1,
								"events": []string{"pty-output"},
							},
						},
					},
				}

				if _, err := c.Database("test").Collection("sessions").InsertMany(ctx, []any{sessions[0]}); err != nil {
					return err
				}

				return nil
			},
			verify: func(t *testing.T) {
				var session struct {
					UID    string               `bson:"uid"`
					Events models.SessionEvents `bson:"events"`
					Seats  []models.SessionSeat `bson:"seats"`
				}

				err := c.Database("test").Collection("sessions").FindOne(ctx, bson.M{"uid": "session-1"}).Decode(&session)
				assert.NoError(t, err)

				assert.ElementsMatch(t, []string{"pty-output", "window-change"}, session.Events.Types)
				assert.ElementsMatch(t, []int{0, 1}, session.Events.Seats)

				assert.Len(t, session.Seats, 0, "Seats should be removed")
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			t.Cleanup(func() { assert.NoError(t, srv.Reset()) })

			require.NoError(t, tc.setup())
			migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[107])
			require.NoError(t, migrates.Up(ctx, migrate.AllAvailable))
			require.NoError(t, migrates.Down(ctx, migrate.AllAvailable))

			tc.verify(t)
		})
	}
}
