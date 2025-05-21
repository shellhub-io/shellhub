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

func TestMigration22(t *testing.T) {
	t.Cleanup(func() {
		assert.NoError(t, srv.Reset())
	})

	user := models.User{
		ID: "1",
	}

	type NamespaceSettings struct {
		SessionRecord bool `json:"session_record" bson:"session_record,omitempty"`
	}

	type Namespace struct {
		Name         string             `json:"name"  validate:"required,hostname_rfc1123,excludes=."`
		Owner        string             `json:"owner"`
		TenantID     string             `json:"tenant_id" bson:"tenant_id,omitempty"`
		Members      []interface{}      `json:"members" bson:"members"`
		Settings     *NamespaceSettings `json:"settings"`
		Devices      int                `json:"devices" bson:",omitempty"`
		Sessions     int                `json:"sessions" bson:",omitempty"`
		MaxDevices   int                `json:"max_devices" bson:"max_devices"`
		DevicesCount int                `json:"devices_count" bson:"devices_count,omitempty"`
		CreatedAt    time.Time          `json:"created_at" bson:"created_at"`
	}

	ns := Namespace{
		Name:       "namespace",
		Owner:      "60df59bc65f88d92b974a60f",
		TenantID:   "tenant",
		Members:    []interface{}{"60df59bc65f88d92b974a60f"},
		MaxDevices: -1,
	}
	_, err := c.Database("test").Collection("devices").InsertOne(context.TODO(), user)
	assert.NoError(t, err)

	_, err = c.Database("test").Collection("namespaces").InsertOne(context.TODO(), ns)
	assert.NoError(t, err)

	migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[21:22]...)
	err = migrates.Up(context.Background(), migrate.AllAvailable)
	assert.NoError(t, err)

	var migratedNamespace *models.Namespace
	err = c.Database("test").Collection("namespaces").FindOne(context.TODO(), bson.M{"tenant_id": "tenant"}).Decode(&migratedNamespace)
	assert.NoError(t, err)
}
