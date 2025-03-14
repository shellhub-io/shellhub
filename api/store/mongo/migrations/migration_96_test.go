package migrations

import (
	"context"
	"slices"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/envs"
	envmock "github.com/shellhub-io/shellhub/pkg/envs/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
)

func TestMigration96Up(t *testing.T) {
	ctx := context.Background()
	mock := &envmock.Backend{}
	envs.DefaultBackend = mock

	cases := []struct {
		description string
		setup       func() error
		verify      func(tt *testing.T)
	}{
		{
			description: "Drop recorded_sessions collection in enterprise mode",
			setup: func() error {
				mock.On("Get", "SHELLHUB_ENTERPRISE").Return("true").Once()

				_, err := c.
					Database("test").
					Collection("sessions").
					InsertOne(ctx, bson.M{
						"uid":           "session-1",
						"authenticated": true,
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
							"message": "test output 1",
							"time":    "2023-01-01T10:00:00Z",
							"width":   80,
							"height":  24,
						},
						bson.M{
							"uid":     "session-1",
							"message": "test output 2",
							"time":    "2023-01-01T10:01:00Z",
							"width":   80,
							"height":  24,
						},
					})

				return err
			},
			verify: func(tt *testing.T) {
				collections, err := c.Database("test").ListCollectionNames(ctx, bson.M{})
				require.NoError(tt, err)

				hasRecordedSessions := slices.Contains(collections, "recorded_sessions")

				if hasRecordedSessions {
					count, err := c.
						Database("test").
						Collection("recorded_sessions").
						CountDocuments(ctx, bson.M{})
					require.NoError(tt, err)
					assert.Equal(tt, int64(0), count)
				}

				count, err := c.
					Database("test").
					Collection("sessions").
					CountDocuments(ctx, bson.M{})
				require.NoError(tt, err)
				assert.Equal(tt, int64(1), count)
			},
		},
		{
			description: "Nothing happens when not in enterprise mode",
			setup: func() error {
				mock.On("Get", "SHELLHUB_ENTERPRISE").Return("false").Once()

				_, err := c.
					Database("test").
					Collection("sessions").
					InsertOne(ctx, bson.M{
						"uid":           "session-2",
						"authenticated": true,
					})
				if err != nil {
					return err
				}

				_, err = c.
					Database("test").
					Collection("recorded_sessions").
					InsertOne(ctx, bson.M{
						"uid":     "session-2",
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
					CountDocuments(ctx, bson.M{"uid": "session-2"})
				require.NoError(tt, err)
				assert.Equal(tt, int64(1), count)
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

func TestMigration96Down(t *testing.T) {
	ctx := context.Background()
	mock := &envmock.Backend{}
	envs.DefaultBackend = mock

	cases := []struct {
		description string
		setup       func() error
		verify      func(tt *testing.T)
	}{
		{
			description: "Migration down has no effect",
			setup: func() error {
				mock.On("Get", "SHELLHUB_ENTERPRISE").Return("true").Once()

				_, err := c.
					Database("test").
					Collection("sessions").
					InsertOne(ctx, bson.M{
						"uid":           "session-3",
						"authenticated": true,
					})
				if err != nil {
					return err
				}

				migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[94])
				if err := migrates.Up(ctx, migrate.AllAvailable); err != nil {
					return err
				}

				return nil
			},
			verify: func(tt *testing.T) {
				collections, err := c.Database("test").ListCollectionNames(ctx, bson.M{})
				require.NoError(tt, err)

				hasRecordedSessions := slices.Contains(collections, "recorded_sessions")

				if hasRecordedSessions {
					count, err := c.
						Database("test").
						Collection("recorded_sessions").
						CountDocuments(ctx, bson.M{})
					require.NoError(tt, err)
					assert.Equal(tt, int64(0), count)
				}

				count, err := c.
					Database("test").
					Collection("sessions").
					CountDocuments(ctx, bson.M{})
				require.NoError(tt, err)
				assert.Equal(tt, int64(1), count)
			},
		},
		{
			description: "Nothing happens when not in enterprise mode",
			setup: func() error {
				mock.On("Get", "SHELLHUB_ENTERPRISE").Return("false").Twice()

				_, err := c.
					Database("test").
					Collection("sessions").
					InsertOne(ctx, bson.M{
						"uid":           "session-4",
						"authenticated": true,
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
				if err != nil {
					return err
				}

				migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[95])
				if err := migrates.Up(ctx, migrate.AllAvailable); err != nil {
					return err
				}

				return nil
			},
			verify: func(tt *testing.T) {
				count, err := c.
					Database("test").
					Collection("recorded_sessions").
					CountDocuments(ctx, bson.M{"uid": "session-4"})
				require.NoError(tt, err)
				assert.Equal(tt, int64(1), count)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			tt.Cleanup(func() {
				assert.NoError(tt, srv.Reset())
			})

			require.NoError(tt, tc.setup())

			migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[95])
			require.NoError(tt, migrates.Down(ctx, migrate.AllAvailable))

			tc.verify(tt)
		})
	}
}
