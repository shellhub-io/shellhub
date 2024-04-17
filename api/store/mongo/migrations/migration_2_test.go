package migrations

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/api/pkg/fixtures"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
)

func TestMigration2(t *testing.T) {
	t.Cleanup(func() {
		assert.NoError(t, fixtures.Teardown())
	})

	type Session struct {
		UID       string `json:"uid"`
		DeviceUID string `json:"device,omitempty" bson:"device"`
		TenantID  string `json:"tenant_id" bson:"tenant_id"`
		Username  string `json:"username"`
		IPAddress string `json:"ip_address" bson:"ip_address"`
	}

	session := Session{
		Username:  "user",
		UID:       "uid",
		DeviceUID: "deviceUID",
		IPAddress: "0.0.0.0",
	}

	_, err := srv.Client().Database("test").Collection("sessions").InsertOne(context.TODO(), session)
	assert.NoError(t, err)

	var afterMigrationSession *Session
	err = srv.Client().Database("test").Collection("sessions").FindOne(context.TODO(), bson.M{"device": "deviceUID"}).Decode(&afterMigrationSession)
	assert.NoError(t, err)

	migrates := migrate.NewMigrate(srv.Client().Database("test"), GenerateMigrations()[:2]...)
	err = migrates.Up(context.Background(), migrate.AllAvailable)
	assert.NoError(t, err)

	_, err = srv.Client().Database("test").Collection("sessions").InsertOne(context.TODO(), session)
	assert.NoError(t, err)

	var migratedSession *models.Session
	err = srv.Client().Database("test").Collection("sessions").FindOne(context.TODO(), bson.M{"device_uid": "deviceUID"}).Decode(&migratedSession)
	assert.NoError(t, err)
}
