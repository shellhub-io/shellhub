package migrations

import (
	"context"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	migrate "github.com/xakep666/mongo-migrate"
)

func TestMigration13(t *testing.T) {
	type ConnectedDevice struct {
		UID      string    `json:"uid"`
		TenantID string    `json:"tenant_id" bson:"tenant_id"`
		LastSeen time.Time `json:"last_seen" bson:"last_seen"`
	}

	t.Cleanup(func() {
		assert.NoError(t, srv.Reset())
	})

	logrus.Info("Test if the UID is unique in the devices collection")

	device1 := models.Device{
		UID: "1",
	}

	device2 := models.Device{
		UID: "1",
	}

	_, err := c.Database("test").Collection("devices").InsertOne(context.TODO(), device1)
	assert.NoError(t, err)

	_, err = c.Database("test").Collection("devices").InsertOne(context.TODO(), device2)
	assert.NoError(t, err)

	migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[:13]...)
	err = migrates.Up(context.Background(), migrate.AllAvailable)
	assert.Error(t, err)

	logrus.Info("Test if the uid in the connected_devices collection is not unique")

	connectedDevice1 := ConnectedDevice{
		UID: "1",
	}

	connectedDevice2 := ConnectedDevice{
		UID: "1",
	}

	_, err = c.Database("test").Collection("connected_devices").InsertOne(context.TODO(), connectedDevice1)
	assert.NoError(t, err)

	_, err = c.Database("test").Collection("connected_devices").InsertOne(context.TODO(), connectedDevice2)
	assert.NoError(t, err)

	logrus.Info("Test if the uid in the sessions collection is unique")

	session1 := models.Session{
		UID: "1",
	}

	session2 := models.Session{
		UID: "1",
	}

	_, err = c.Database("test").Collection("sessions").InsertOne(context.TODO(), session1)
	assert.NoError(t, err)

	_, err = c.Database("test").Collection("sessions").InsertOne(context.TODO(), session2)
	assert.NoError(t, err)

	activeSession1 := models.ActiveSession{
		UID: "1",
	}

	activeSession2 := models.ActiveSession{
		UID: "1",
	}

	_, err = c.Database("test").Collection("active_sessions").InsertOne(context.TODO(), activeSession1)
	assert.NoError(t, err)

	_, err = c.Database("test").Collection("active_sessions").InsertOne(context.TODO(), activeSession2)
	assert.NoError(t, err)

	logrus.Info("Test if the tenant_id in the users collection is unique")

	type user struct {
		Username      string `json:"username" bson:",omitempty"`
		Email         string `json:"email" bson:",omitempty" validate:"required,email"`
		TenantID      string `json:"tenant_id" bson:"tenant_id"`
		ID            string `json:"id,omitempty" bson:"_id,omitempty"`
		SessionRecord bool   `json:"session_record" bson:"session_record,omitempty"`
	}

	user1 := user{
		TenantID: "1",
		Email:    "test1",
	}

	user2 := user{
		TenantID: "1",
		Email:    "test2",
	}

	_, err = c.Database("test").Collection("users").InsertOne(context.TODO(), user1)
	assert.NoError(t, err)

	_, err = c.Database("test").Collection("users").InsertOne(context.TODO(), user2)
	assert.NoError(t, err)
}
