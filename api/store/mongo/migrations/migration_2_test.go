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

func TestMigration2(t *testing.T) {
	logrus.Info("Testing Migration 2 - Test if the column device was renamed to device_uid")

	db := dbtest.DBServer{}
	defer db.Stop()

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

	_, err := db.Client().Database("test").Collection("sessions").InsertOne(context.TODO(), session)
	assert.NoError(t, err)

	var afterMigrationSession *Session
	err = db.Client().Database("test").Collection("sessions").FindOne(context.TODO(), bson.M{"device": "deviceUID"}).Decode(&afterMigrationSession)
	assert.NoError(t, err)

	migrates := migrate.NewMigrate(db.Client().Database("test"), GenerateMigrations()[:2]...)
	err = migrates.Up(migrate.AllAvailable)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("sessions").InsertOne(context.TODO(), session)
	assert.NoError(t, err)

	var migratedSession *models.Session
	err = db.Client().Database("test").Collection("sessions").FindOne(context.TODO(), bson.M{"device_uid": "deviceUID"}).Decode(&migratedSession)
	assert.NoError(t, err)
}
