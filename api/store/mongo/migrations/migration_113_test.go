package migrations

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
)

func TestMigration113Up(t *testing.T) {
	ctx := context.Background()

	cases := []struct {
		description string
		setup       func() error
		verify      func(tt *testing.T)
	}{
		{
			description: "succeeds adding tls structure with enabled, verify, and domain fields to all web_endpoints",
			setup: func() error {
				endpoints := []bson.M{
					{"address": "endpoint1", "host": "192.168.1.1", "port": 8080},
					{"address": "endpoint2", "host": "192.168.1.2", "port": 8081},
				}
				_, err := c.Database("test").Collection("tunnels").InsertMany(ctx, []any{endpoints[0], endpoints[1]})

				return err
			},
			verify: func(tt *testing.T) {
				cursor, err := c.Database("test").Collection("tunnels").Find(ctx, bson.M{})
				require.NoError(tt, err)

				var results []bson.M
				require.NoError(tt, cursor.All(ctx, &results))

				for _, endpoint := range results {
					tls, tlsExists := endpoint["tls"]
					assert.True(tt, tlsExists)

					tlsObj, ok := tls.(bson.M)
					require.True(tt, ok)

					enabled, enabledExists := tlsObj["enabled"]
					assert.True(tt, enabledExists)
					assert.Equal(tt, false, enabled)

					verify, verifyExists := tlsObj["verify"]
					assert.True(tt, verifyExists)
					assert.Equal(tt, false, verify)

					domain, domainExists := tlsObj["domain"]
					assert.True(tt, domainExists)
					assert.Equal(tt, "", domain)
				}
			},
		},
		{
			description: "succeeds when web_endpoints collection is empty",
			setup: func() error {
				return nil
			},
			verify: func(tt *testing.T) {
				cursor, err := c.Database("test").Collection("tunnels").Find(ctx, bson.M{})
				require.NoError(tt, err)

				var results []bson.M
				require.NoError(tt, cursor.All(ctx, &results))

				assert.Equal(tt, 0, len(results))
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			tt.Cleanup(func() { assert.NoError(tt, srv.Reset()) })

			require.NoError(tt, tc.setup())
			migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[112]) // migration113
			require.NoError(tt, migrates.Up(ctx, migrate.AllAvailable))
			tc.verify(tt)
		})
	}
}

func TestMigration113Down(t *testing.T) {
	ctx := context.Background()

	cases := []struct {
		description string
		setup       func() error
		verify      func(tt *testing.T)
	}{
		{
			description: "succeeds removing tls structure from all web_endpoints",
			setup: func() error {
				endpoints := []bson.M{
					{"address": "endpoint1", "host": "192.168.1.1", "port": 8080, "tls": bson.M{"enabled": true, "verify": true, "domain": "example.com"}},
					{"address": "endpoint2", "host": "192.168.1.2", "port": 8081, "tls": bson.M{"enabled": false, "verify": false, "domain": "test.com"}},
				}
				_, err := c.Database("test").Collection("tunnels").InsertMany(ctx, []any{endpoints[0], endpoints[1]})

				return err
			},
			verify: func(tt *testing.T) {
				cursor, err := c.Database("test").Collection("tunnels").Find(ctx, bson.M{})
				require.NoError(tt, err)

				var results []bson.M
				require.NoError(tt, cursor.All(ctx, &results))

				for _, endpoint := range results {
					_, tlsExists := endpoint["tls"]
					assert.False(tt, tlsExists)
				}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			tt.Cleanup(func() { assert.NoError(tt, srv.Reset()) })

			require.NoError(tt, tc.setup())
			migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[112])
			require.NoError(tt, migrates.Up(ctx, migrate.AllAvailable))
			require.NoError(tt, migrates.Down(ctx, migrate.AllAvailable))
			tc.verify(tt)
		})
	}
}
