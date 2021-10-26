package migrations

import (
	"context"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/shellhub-io/shellhub/pkg/authorizer"
)

func TestMigration37(t *testing.T) {
	logrus.Info("Testing Migration 37")

	db := dbtest.DBServer{}
	defer db.Stop()

	user := models.User{
		ID: "60df59bc65f88d92b974a60f",
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
		Name:     "userspace",
		Owner:    user.ID,
		TenantID: "tenant",
		Members:  []interface{}{user.ID},
		Devices:  -1,
	}
	migrations := GenerateMigrations()[36:37]

	_, err := db.Client().Database("test").Collection("namespaces").InsertOne(context.TODO(), ns)
	assert.NoError(t, err)

	migrates := migrate.NewMigrate(db.Client().Database("test"), migrations...)
	err = migrates.Up(migrate.AllAvailable)
	assert.NoError(t, err)

	migratedNamespace := &models.Namespace{}
	err = db.Client().Database("test").Collection("namespaces").FindOne(context.TODO(), bson.D{{"tenant_id", "tenant"}}).Decode(migratedNamespace)
	assert.NoError(t, err)
	assert.Equal(t, []models.Member{{ID: user.ID, Type: authorizer.MemberTypeOwner}}, migratedNamespace.Members)

	namespace := models.Namespace{
		Name:     "userspace",
		Owner:    user.ID,
		TenantID: "tenant",
		Members:  []models.Member{{ID: user.ID, Type: authorizer.MemberTypeOwner}},
		Devices:  -1,
	}

	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(context.TODO(), namespace)
	assert.NoError(t, err)
	migrates = migrate.NewMigrate(db.Client().Database("test"), migrations...)
	err = migrates.Down(migrate.AllAvailable)
	assert.NoError(t, err)

	migratedNamespaceDown := &Namespace{}
	err = db.Client().Database("test").Collection("namespaces").FindOne(context.TODO(), bson.D{{"tenant_id", namespace.TenantID}}).Decode(migratedNamespaceDown)
	assert.NoError(t, err)
	assert.Equal(t, []interface{}{user.ID}, migratedNamespaceDown.Members)
}
