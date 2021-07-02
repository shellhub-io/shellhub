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
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestMigration20(t *testing.T) {
	logrus.Info("Testing Migration 20 - Test if the firewall_rules has change to new one")

	db := dbtest.DBServer{}
	defer db.Stop()

	type firewallRule struct {
		ID                        primitive.ObjectID `json:"id" bson:"_id"`
		TenantID                  string             `json:"tenant_id" bson:"tenant_id"`
		models.FirewallRuleFields `bson:",inline"`
	}

	fRule := firewallRule{
		TenantID: "tenant",
	}

	_, err := db.Client().Database("test").Collection("firewall_rules").InsertOne(context.TODO(), fRule)
	assert.NoError(t, err)

	migrates := migrate.NewMigrate(db.Client().Database("test"), GenerateMigrations()[19:20]...)
	err = migrates.Up(migrate.AllAvailable)
	assert.NoError(t, err)

	var migratedFirewallRules *models.FirewallRule
	err = db.Client().Database("test").Collection("firewall_rules").FindOne(context.TODO(), bson.M{"tenant_id": "tenant"}).Decode(&migratedFirewallRules)
	assert.NoError(t, err)
}
