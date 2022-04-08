package migrations

import (
	"context"
	"sort"
	"testing"

	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
)

func TestMigration45(t *testing.T) {
	logrus.Info("Testing Migration 45")

	db := dbtest.DBServer{}
	defer db.Stop()

	ruleTagDuplicated := &models.FirewallRule{
		ID:       "id",
		TenantID: "tenant",
		FirewallRuleFields: models.FirewallRuleFields{
			Priority: 1,
			Action:   "",
			Active:   true,
			SourceIP: ".*",
			Username: ".*",
			Filter: models.FirewallFilter{
				Tags: []string{"tag1", "tag2", "tag2"},
			},
		},
	}

	ruleTagWithoutDuplication := &models.FirewallRule{
		ID:       "id",
		TenantID: "tenant",
		FirewallRuleFields: models.FirewallRuleFields{
			Priority: 1,
			Action:   "",
			Active:   true,
			SourceIP: ".*",
			Username: ".*",
			Filter: models.FirewallFilter{
				Tags: []string{"tag1", "tag2"},
			},
		},
	}

	ruleTagNoDuplicated := &models.FirewallRule{
		ID:       "id1",
		TenantID: "tenant1",
		FirewallRuleFields: models.FirewallRuleFields{
			Priority: 1,
			Action:   "",
			Active:   true,
			SourceIP: ".*",
			Username: ".*",
			Filter: models.FirewallFilter{
				Tags: []string{"tag1", "tag3"},
			},
		},
	}

	ruleHostname := &models.FirewallRule{
		ID:       "id2",
		TenantID: "tenant2",
		FirewallRuleFields: models.FirewallRuleFields{
			Priority: 1,
			Action:   "",
			Active:   true,
			SourceIP: ".*",
			Username: ".*",
			Filter: models.FirewallFilter{
				Hostname: ".*",
			},
		},
	}

	_, err := db.Client().Database("test").Collection("firewall_rules").InsertOne(context.TODO(), ruleTagDuplicated)
	assert.NoError(t, err)
	_, err = db.Client().Database("test").Collection("firewall_rules").InsertOne(context.TODO(), ruleTagNoDuplicated)
	assert.NoError(t, err)
	_, err = db.Client().Database("test").Collection("firewall_rules").InsertOne(context.TODO(), ruleHostname)
	assert.NoError(t, err)

	cases := []struct {
		description string
		Test        func(t *testing.T)
	}{
		{
			"Success to apply up on migration 45 when firewall rule tags are duplicated",
			func(t *testing.T) {
				t.Helper()

				migrations := GenerateMigrations()[44:45]
				migrates := migrate.NewMigrate(db.Client().Database("test"), migrations...)
				err := migrates.Up(migrate.AllAvailable)
				assert.NoError(t, err)

				rule := new(models.FirewallRule)
				result := db.Client().Database("test").Collection("firewall_rules").FindOne(context.TODO(), bson.M{"tenant_id": ruleTagDuplicated.TenantID})
				assert.NoError(t, result.Err())

				err = result.Decode(rule)
				assert.NoError(t, err)

				sort.Strings(rule.Filter.Tags)

				assert.Equal(t, ruleTagWithoutDuplication, rule)
			},
		},
		{
			"Success to apply up on migration 45 when firewall rule tags are not duplicated",
			func(t *testing.T) {
				t.Helper()

				migrations := GenerateMigrations()[44:45]
				migrates := migrate.NewMigrate(db.Client().Database("test"), migrations...)
				err := migrates.Up(migrate.AllAvailable)
				assert.NoError(t, err)

				rule := new(models.FirewallRule)
				result := db.Client().Database("test").Collection("firewall_rules").FindOne(context.TODO(), bson.M{"tenant_id": ruleTagNoDuplicated.TenantID})
				assert.NoError(t, result.Err())

				err = result.Decode(rule)
				assert.NoError(t, err)

				sort.Strings(rule.Filter.Tags)

				assert.Equal(t, ruleTagNoDuplicated, rule)
			},
		},
		{
			"Success to apply up on migration 45 when firewall rule has hostname",
			func(t *testing.T) {
				t.Helper()

				migrations := GenerateMigrations()[44:45]
				migrates := migrate.NewMigrate(db.Client().Database("test"), migrations...)
				err := migrates.Up(migrate.AllAvailable)
				assert.NoError(t, err)

				rule := new(models.FirewallRule)
				result := db.Client().Database("test").Collection("firewall_rules").FindOne(context.TODO(), bson.M{"tenant_id": ruleHostname.TenantID})
				assert.NoError(t, result.Err())

				err = result.Decode(rule)
				assert.NoError(t, err)

				assert.Equal(t, ruleHostname, rule)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, tc.Test)
	}
}
