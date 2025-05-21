package migrations

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
)

func TestMigration36(t *testing.T) {
	t.Cleanup(func() {
		assert.NoError(t, srv.Reset())
	})

	migrations := GenerateMigrations()[:35]

	migrates := migrate.NewMigrate(c.Database("test"), migrations...)
	err := migrates.Up(context.Background(), migrate.AllAvailable)
	assert.NoError(t, err)

	version, _, err := migrates.Version(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, uint64(35), version)

	cases := []struct {
		description       string
		isCloud           bool
		toBeMigrated      models.Namespace
		migratedNamespace models.Namespace
		expected          int
	}{
		{
			description: "migrate cloud instance",
			isCloud:     true,
			toBeMigrated: models.Namespace{
				Name:       "ns1",
				TenantID:   "xxx1",
				MaxDevices: -1,
			},
			expected: 3,
		},
		{
			description: "do not apply migration for cloud disabled",
			isCloud:     false,
			toBeMigrated: models.Namespace{
				Name:       "ns2",
				TenantID:   "xxx2",
				MaxDevices: -1,
			},
			expected: -1,
		},
		{
			description: "avoid update active instance",
			isCloud:     true,
			toBeMigrated: models.Namespace{
				Name:     "ns3",
				TenantID: "xxx3",
				Billing: &models.Billing{
					SubscriptionID: "sub_123",
				},
				MaxDevices: -1,
			},
			expected: -1,
		},
	}

	namespaces := make([]interface{}, len(cases))
	for i, v := range cases {
		namespaces[i] = v.toBeMigrated
	}

	_, err = c.Database("test").Collection("namespaces").InsertMany(context.TODO(), namespaces)
	assert.NoError(t, err)

	migrates = migrate.NewMigrate(c.Database("test"), GenerateMigrations()[35])
	err = migrates.Up(context.Background(), migrate.AllAvailable)
	assert.NoError(t, err)

	version, _, err = migrates.Version(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, uint64(36), version)

	cur, err := c.Database("test").Collection("namespaces").Find(context.TODO(), bson.D{})
	assert.NoError(t, err)

	index := 0

	for cur.Next(context.TODO()) {
		var ns models.Namespace
		err := cur.Decode(&ns)
		if err != nil {
			panic(err.Error())
		}

		cases[index].migratedNamespace = ns
		index++
	}

	instance := envs.IsCloud()

	for _, tc := range cases {
		if instance == tc.isCloud {
			t.Run(tc.description, func(t *testing.T) {
				assert.Equal(t, tc.expected, tc.migratedNamespace.MaxDevices)
			})
		}
	}
}
