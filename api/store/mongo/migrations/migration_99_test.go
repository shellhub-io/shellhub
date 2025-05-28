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
			description: "mark recorded=false for sessions with empty events.types",
			setup: func() error {
				_, err := c.
					Database("test").
					Collection("sessions").
					InsertOne(ctx, bson.M{
						"recorded": true,
						"events":   bson.M{"types": []interface{}{}},
					})

				return err
			},
			verify: func(tt *testing.T) {
				session := make(map[string]any)
				require.NoError(tt, c.Database("test").Collection("sessions").FindOne(ctx, bson.M{}).Decode(&session))

				recorded, ok := session["recorded"]
				require.True(tt, ok)
				require.Equal(tt, false, recorded)
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

func TestMigration99Down(t *testing.T) {
	ctx := context.Background()

	cases := []struct {
		description string
		setup       func() error
		verify      func(tt *testing.T)
	}{
		{
			description: "revert recorded=true for sessions with empty events.types",
			setup: func() error {
				_, err := c.
					Database("test").
					Collection("sessions").
					InsertOne(ctx, bson.M{
						"recorded": false,
						"events":   bson.M{"types": []interface{}{}},
					})

				return err
			},
			verify: func(tt *testing.T) {
				session := make(map[string]any)
				require.NoError(tt, c.Database("test").Collection("sessions").FindOne(ctx, bson.M{}).Decode(&session))

				recorded, ok := session["recorded"]
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
			require.NoError(tt, migrates.Down(ctx, migrate.AllAvailable))

			tc.verify(tt)
		})
	}
}
