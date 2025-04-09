package migrations

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestMigration20(t *testing.T) {
	t.Cleanup(func() {
		assert.NoError(t, srv.Reset())
	})

	type firewallRule struct {
		ID                        primitive.ObjectID `json:"id" bson:"_id"`
		TenantID                  string             `json:"tenant_id" bson:"tenant_id"`
		models.FirewallRuleFields `bson:",inline"`
	}

	fRule := firewallRule{
		TenantID: "tenant",
	}

	_, err := c.Database("test").Collection("firewall_rules").InsertOne(context.TODO(), fRule)
	assert.NoError(t, err)

	migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[19:20]...)
	err = migrates.Up(context.Background(), migrate.AllAvailable)
	assert.NoError(t, err)

	var migratedFirewallRules *models.FirewallRule
	err = c.Database("test").Collection("firewall_rules").FindOne(context.TODO(), bson.M{"tenant_id": "tenant"}).Decode(&migratedFirewallRules)
	assert.NoError(t, err)
}
