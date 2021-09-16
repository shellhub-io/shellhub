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

func TestMigration17(t *testing.T) {
	logrus.Info("Testing Migration 17 - Test if the namespaces, devices, session, connected_devices, firewall_rules and public_keys was deleted for users")

	db := dbtest.DBServer{}
	defer db.Stop()

	user := models.User{
		Name:     "name",
		Username: "username",
		Password: "password",
		Email:    "email",
	}

	namespace := models.Namespace{
		Name:     "name",
		Owner:    "60df59bc65f88d92b974a60f",
		TenantID: "tenant",
	}

	device := models.Device{
		UID:      "1",
		TenantID: "tenant",
	}

	session := models.Session{
		DeviceUID: "1",
	}

	connectedDevice := connectedDevice{
		UID: "1",
	}

	firewallRules := models.FirewallRule{
		TenantID: "tenant",
	}

	pk := models.PublicKey{
		TenantID: "tenant",
	}

	_, err := db.Client().Database("test").Collection("users").InsertOne(context.TODO(), user)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(context.TODO(), namespace)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("devices").InsertOne(context.TODO(), device)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("sessions").InsertOne(context.TODO(), session)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("connected_devices").InsertOne(context.TODO(), connectedDevice)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("firewall_rules").InsertOne(context.TODO(), firewallRules)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("public_keys").InsertOne(context.TODO(), pk)
	assert.NoError(t, err)

	migrates := migrate.NewMigrate(db.Client().Database("test"), GenerateMigrations()[:17]...)
	err = migrates.Up(migrate.AllAvailable)
	assert.NoError(t, err)

	err = db.Client().Database("test").Collection("namespaces").FindOne(context.TODO(), bson.M{"tenant_id": namespace.TenantID}).Decode(&namespace)
	assert.Error(t, err)

	err = db.Client().Database("test").Collection("devices").FindOne(context.TODO(), bson.M{"tenant_id": device.TenantID}).Decode(&device)
	assert.Error(t, err)

	err = db.Client().Database("test").Collection("sessions").FindOne(context.TODO(), bson.M{"device_uid": session.DeviceUID}).Decode(&session)
	assert.Error(t, err)

	err = db.Client().Database("test").Collection("connected_devices").FindOne(context.TODO(), bson.M{"uid": connectedDevice.UID}).Decode(&connectedDevice)
	assert.Error(t, err)

	err = db.Client().Database("test").Collection("firewall_rules").FindOne(context.TODO(), bson.M{"tenant_id": firewallRules.TenantID}).Decode(&firewallRules)
	assert.Error(t, err)

	err = db.Client().Database("test").Collection("public_keys").FindOne(context.TODO(), bson.M{"tenant_id": pk.TenantID}).Decode(&pk)
	assert.Error(t, err)
}
