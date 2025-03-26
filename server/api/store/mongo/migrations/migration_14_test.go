package migrations

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
)

func TestMigration14(t *testing.T) {
	t.Cleanup(func() {
		assert.NoError(t, srv.Reset())
	})

	type user struct {
		Username      string `json:"username" bson:",omitempty"`
		TenantID      string `json:"tenant_id" bson:"tenant_id"`
		ID            string `json:"id,omitempty" bson:"_id,omitempty"`
		SessionRecord bool   `json:"session_record" bson:"session_record,omitempty"`
	}

	user1 := user{
		TenantID: "1",
		ID:       "1",
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
		Owner:    "1",
		TenantID: "1",
	}

	_, err := c.Database("test").Collection("users").InsertOne(context.TODO(), user1)
	assert.NoError(t, err)

	_, err = c.Database("test").Collection("namespaces").InsertOne(context.TODO(), ns)
	assert.NoError(t, err)

	migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[:14]...)
	err = migrates.Up(context.Background(), migrate.AllAvailable)
	assert.NoError(t, err)

	err = c.Database("test").Collection("users").FindOne(context.TODO(), bson.M{"tenant_id": "1"}).Decode(&user1)
	assert.NoError(t, err)
}
