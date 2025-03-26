package migrations

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestMigration41(t *testing.T) {
	cases := []struct {
		description string
		Test        func(t *testing.T)
	}{
		{
			"Success to apply up on migration 41",
			func(t *testing.T) {
				t.Helper()

				oldIndex := mongo.IndexModel{
					Keys:    bson.D{{"last_seen", 1}},
					Options: options.Index().SetName("last_seen").SetExpireAfterSeconds(60),
				}
				newIndex := mongo.IndexModel{
					Keys:    bson.D{{"last_seen", 1}},
					Options: options.Index().SetName("last_seen").SetExpireAfterSeconds(60),
				}
				_, err := c.Database("test").Collection("connected_devices").Indexes().CreateOne(context.TODO(), oldIndex)
				assert.NoError(t, err)

				migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[40:41]...)
				assert.NoError(t, migrates.Up(context.Background(), migrate.AllAvailable))

				_, err = c.Database("test").Collection("connected_devices").Indexes().DropOne(context.TODO(), "last_seen")
				assert.NoError(t, err)

				_, err = c.Database("test").Collection("connected_devices").Indexes().CreateOne(context.TODO(), newIndex)
				assert.NoError(t, err)

				const Expected = 1
				list, err := c.Database("test").Collection("connected_devices").Indexes().ListSpecifications(context.TODO())
				assert.NoError(t, err)

				assert.Equal(t, newIndex.Options.ExpireAfterSeconds, list[Expected].ExpireAfterSeconds)
			},
		},
		{
			"Success to apply down on migration 41",
			func(t *testing.T) {
				t.Helper()

				oldIndex := mongo.IndexModel{
					Keys:    bson.D{{"last_seen", 1}},
					Options: options.Index().SetName("last_seen").SetExpireAfterSeconds(60),
				}
				_, err := c.Database("test").Collection("connected_devices").Indexes().CreateOne(context.TODO(), oldIndex)
				assert.NoError(t, err)

				migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[40:41]...)
				assert.NoError(t, migrates.Down(context.Background(), migrate.AllAvailable))

				assert.NoError(t, err)
				_, err = c.Database("test").Collection("connected_devices").Indexes().DropOne(context.TODO(), "last_seen")
				assert.NoError(t, err)

				_, err = c.Database("test").Collection("connected_devices").Indexes().CreateOne(context.TODO(), oldIndex)
				assert.NoError(t, err)

				const Expected = 1
				list, err := c.Database("test").Collection("connected_devices").Indexes().ListSpecifications(context.TODO())
				assert.NoError(t, err)

				assert.Equal(t, oldIndex.Options.ExpireAfterSeconds, list[Expected].ExpireAfterSeconds)
			},
		},
	}

	for _, test := range cases {
		tc := test
		t.Run(tc.description, func(t *testing.T) {
			t.Cleanup(func() {
				assert.NoError(t, srv.Reset())
			})
			tc.Test(t)
		})
	}
}
