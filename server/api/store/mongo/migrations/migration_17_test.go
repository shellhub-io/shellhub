package migrations

import (
	"context"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
)

func TestMigration17(t *testing.T) {
	t.Cleanup(func() {
		assert.NoError(t, srv.Reset())
	})

	user := models.User{
		UserData: models.UserData{
			Name:     "name",
			Username: "username",
			Email:    "email",
		},
		Password: models.UserPassword{
			Hash: "2bb80d537b1da3e38bd30361aa855686bde0eacd7162fef6a25fe97bf527a25b",
		},
	}

	type NamespaceSettings struct {
		SessionRecord bool `json:"session_record" bson:"session_record,omitempty"`
	}

	type Member struct {
		ID   string `json:"id" bson:"id"`
		Name string `json:"name,omitempty" bson:"-"`
	}

	type ConnectedDevice struct {
		UID      string    `json:"uid"`
		TenantID string    `json:"tenant_id" bson:"tenant_id"`
		LastSeen time.Time `json:"last_seen" bson:"last_seen"`
	}

	type Namespace struct {
		Name         string             `json:"name"  validate:"required,hostname_rfc1123,excludes=."`
		Owner        string             `json:"owner"`
		TenantID     string             `json:"tenant_id" bson:"tenant_id,omitempty"`
		Members      []Member           `json:"members" bson:"members"`
		Settings     *NamespaceSettings `json:"settings"`
		Devices      int                `json:"devices" bson:",omitempty"`
		Sessions     int                `json:"sessions" bson:",omitempty"`
		MaxDevices   int                `json:"max_devices" bson:"max_devices"`
		DevicesCount int                `json:"devices_count" bson:"devices_count,omitempty"`
		CreatedAt    time.Time          `json:"created_at" bson:"created_at"`
	}

	namespace := Namespace{
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

	connectedDevice := ConnectedDevice{
		UID: "1",
	}

	firewallRules := models.FirewallRule{
		TenantID: "tenant",
	}

	pk := models.PublicKey{
		TenantID: "tenant",
	}

	_, err := c.Database("test").Collection("users").InsertOne(context.TODO(), user)
	assert.NoError(t, err)

	_, err = c.Database("test").Collection("namespaces").InsertOne(context.TODO(), namespace)
	assert.NoError(t, err)

	_, err = c.Database("test").Collection("devices").InsertOne(context.TODO(), device)
	assert.NoError(t, err)

	_, err = c.Database("test").Collection("sessions").InsertOne(context.TODO(), session)
	assert.NoError(t, err)

	_, err = c.Database("test").Collection("connected_devices").InsertOne(context.TODO(), connectedDevice)
	assert.NoError(t, err)

	_, err = c.Database("test").Collection("firewall_rules").InsertOne(context.TODO(), firewallRules)
	assert.NoError(t, err)

	_, err = c.Database("test").Collection("public_keys").InsertOne(context.TODO(), pk)
	assert.NoError(t, err)

	migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[:17]...)
	err = migrates.Up(context.Background(), migrate.AllAvailable)
	assert.NoError(t, err)

	err = c.Database("test").Collection("namespaces").FindOne(context.TODO(), bson.M{"tenant_id": namespace.TenantID}).Decode(&namespace)
	assert.Error(t, err)

	err = c.Database("test").Collection("devices").FindOne(context.TODO(), bson.M{"tenant_id": device.TenantID}).Decode(&device)
	assert.Error(t, err)

	err = c.Database("test").Collection("sessions").FindOne(context.TODO(), bson.M{"device_uid": session.DeviceUID}).Decode(&session)
	assert.Error(t, err)

	err = c.Database("test").Collection("connected_devices").FindOne(context.TODO(), bson.M{"uid": connectedDevice.UID}).Decode(&connectedDevice)
	assert.Error(t, err)

	err = c.Database("test").Collection("firewall_rules").FindOne(context.TODO(), bson.M{"tenant_id": firewallRules.TenantID}).Decode(&firewallRules)
	assert.Error(t, err)

	err = c.Database("test").Collection("public_keys").FindOne(context.TODO(), bson.M{"tenant_id": pk.TenantID}).Decode(&pk)
	assert.Error(t, err)
}
