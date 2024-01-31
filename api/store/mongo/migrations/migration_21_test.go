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

func TestMigration21(t *testing.T) {
	logrus.Info("Testing Migration 21 - Test if the sessions and connected_devices was removed for the devices")

	db := dbtest.DB{}
	err := func() error {
		err := db.Down(context.Background())

		return err
	}()
	assert.NoError(t, err)

	device := models.Device{
		UID: "1",
	}

	recordedSession := models.RecordedSession{
		TenantID: "tenant",
	}

	session := models.Session{
		UID: "1",
	}

	_, err = mongoClient.Database("test").Collection("devices").InsertOne(context.TODO(), device)
	assert.NoError(t, err)

	_, err = mongoClient.Database("test").Collection("recorded_sessions").InsertOne(context.TODO(), recordedSession)
	assert.NoError(t, err)

	_, err = mongoClient.Database("test").Collection("sessions").InsertOne(context.TODO(), session)
	assert.NoError(t, err)

	migrates := migrate.NewMigrate(mongoClient.Database("test"), GenerateMigrations()[:21]...)
	err = migrates.Up(context.Background(), migrate.AllAvailable)
	assert.NoError(t, err)

	var migratedRecordedSession *models.RecordedSession
	err = mongoClient.Database("test").Collection("recorded_sessions").FindOne(context.TODO(), bson.M{"tenant_id": "tenant"}).Decode(&migratedRecordedSession)
	assert.Error(t, err)

	var migratedSession *models.Session
	err = mongoClient.Database("test").Collection("sessions").FindOne(context.TODO(), bson.M{"tenant_id": "tenant"}).Decode(&migratedSession)
	assert.Error(t, err)
}
