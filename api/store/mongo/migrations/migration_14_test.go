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

func TestMigration14(t *testing.T) {
	logrus.Info("Testing Migration 14 - Test if the right tenant_id is set")

	db := dbtest.DBServer{}
	defer db.Stop()

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

	ns := models.Namespace{
		Owner:    "1",
		TenantID: "1",
	}

	_, err := db.Client().Database("test").Collection("users").InsertOne(context.TODO(), user1)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(context.TODO(), ns)
	assert.NoError(t, err)

	migrates := migrate.NewMigrate(db.Client().Database("test"), GenerateMigrations()[:14]...)
	err = migrates.Up(migrate.AllAvailable)
	assert.NoError(t, err)

	err = db.Client().Database("test").Collection("users").FindOne(context.TODO(), bson.M{"tenant_id": "1"}).Decode(&user1)
	assert.NoError(t, err)
}
