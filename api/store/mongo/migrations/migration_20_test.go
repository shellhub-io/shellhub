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

	db := dbtest.DB{}
	err := func() error {
		err := db.Down(context.Background())

		return err
	}()
	assert.NoError(t, err)

	type firewallRule struct {
		ID                        primitive.ObjectID `json:"id" bson:"_id"`
		TenantID                  string             `json:"tenant_id" bson:"tenant_id"`
		models.FirewallRuleFields `bson:",inline"`
	}

	fRule := firewallRule{
		TenantID: "tenant",
	}

	_, err = mongoClient.Database("test").Collection("firewall_rules").InsertOne(context.TODO(), fRule)
	assert.NoError(t, err)

	migrates := migrate.NewMigrate(mongoClient.Database("test"), GenerateMigrations()[19:20]...)
	err = migrates.Up(context.Background(), migrate.AllAvailable)
	assert.NoError(t, err)

	var migratedFirewallRules *models.FirewallRule
	err = mongoClient.Database("test").Collection("firewall_rules").FindOne(context.TODO(), bson.M{"tenant_id": "tenant"}).Decode(&migratedFirewallRules)
	assert.NoError(t, err)
}
