package migrations

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
)

func TestMigration48(t *testing.T) {
	logrus.Info("Testing Migration 48")

	ctx := context.Background()

	db := dbtest.DB{}
	defer db.Stop()

	namespace := models.Namespace{
		TenantID: "tenant",
	}

	rule0 := models.FirewallRule{
		TenantID: namespace.TenantID,
		FirewallRuleFields: models.FirewallRuleFields{
			Priority: 0,
		},
	}

	rule1 := models.FirewallRule{
		TenantID: namespace.TenantID,
		FirewallRuleFields: models.FirewallRuleFields{
			Priority: 1,
		},
	}

	rule2 := models.FirewallRule{
		TenantID: namespace.TenantID,
		FirewallRuleFields: models.FirewallRuleFields{
			Priority: 2,
		},
	}

	_, err := db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)
	_, err = db.Client().Database("test").Collection("firewall_rules").InsertOne(ctx, rule0)
	assert.NoError(t, err)
	_, err = db.Client().Database("test").Collection("firewall_rules").InsertOne(ctx, rule1)
	assert.NoError(t, err)
	_, err = db.Client().Database("test").Collection("firewall_rules").InsertOne(ctx, rule2)
	assert.NoError(t, err)

	cases := []struct {
		description string
		Test        func(t *testing.T)
	}{
		{
			"Success to apply up on migration 48",
			func(t *testing.T) {
				t.Helper()

				migrations := GenerateMigrations()[47:48]
				migrates := migrate.NewMigrate(db.Client().Database("test"), migrations...)
				err := migrates.Up(context.Background(), migrate.AllAvailable)
				assert.NoError(t, err)

				key := new(models.FirewallRule)
				result := db.Client().Database("test").Collection("firewall_rules").FindOne(ctx, bson.M{"tenant_id": namespace.TenantID})
				assert.NoError(t, result.Err())

				err = result.Decode(key)
				assert.NoError(t, err)

				assert.Equal(t, 2, key.Priority)
			},
		},
		{
			"Success to apply down on migration 48",
			func(t *testing.T) {
				t.Helper()

				migrations := GenerateMigrations()[47:48]
				migrates := migrate.NewMigrate(db.Client().Database("test"), migrations...)
				err := migrates.Down(context.Background(), migrate.AllAvailable)
				assert.NoError(t, err)

				key := new(models.FirewallRule)
				result := db.Client().Database("test").Collection("firewall_rules").FindOne(ctx, bson.M{"tenant_id": namespace.TenantID})
				assert.NoError(t, result.Err())

				err = result.Decode(key)
				assert.NoError(t, err)

				assert.Equal(t, 0, key.Priority)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, tc.Test)
	}
}
