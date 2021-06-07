package migrations

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestMigration24(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	migrations := GenerateMigrations()[:23]

	migrates := migrate.NewMigrate(db.Client().Database("test"), migrations...)
	err := migrates.Up(migrate.AllAvailable)
	assert.NoError(t, err)

	version, _, err := migrates.Version()
	assert.NoError(t, err)
	assert.Equal(t, uint64(23), version)

	user := models.User{
		Name:     "name",
		Username: "USERNAME",
		Password: "password",
		Email:    "EMAIL@MAIL.COM",
	}
	_, err = db.Client().Database("test").Collection("users").InsertOne(context.TODO(), user)
	assert.NoError(t, err)

	user = models.User{
		Name:     "name2",
		Username: "Username2",
		Password: "password",
		Email:    "email@MAIL-TEST.com",
	}
	_, err = db.Client().Database("test").Collection("users").InsertOne(context.TODO(), user)
	assert.NoError(t, err)

	user = models.User{
		Name:     "name3",
		Username: "username3",
		Password: "password",
		Email:    "email@e-mail.com",
	}
	_, err = db.Client().Database("test").Collection("users").InsertOne(context.TODO(), user)
	assert.NoError(t, err)

	namespace := models.Namespace{
		Name:     "NaMe",
		Owner:    "owner",
		TenantID: "tenant",
	}
	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(context.TODO(), namespace)
	assert.NoError(t, err)

	namespace = models.Namespace{
		Name:     "TEST",
		Owner:    "owner",
		TenantID: "tenant2",
	}
	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(context.TODO(), namespace)
	assert.NoError(t, err)

	namespace = models.Namespace{
		Name:     "teste",
		Owner:    "owner",
		TenantID: "tenant3",
	}
	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(context.TODO(), namespace)
	assert.NoError(t, err)

	migration := GenerateMigrations()[23]

	migrates = migrate.NewMigrate(db.Client().Database("test"), migration)
	err = migrates.Up(migrate.AllAvailable)
	assert.NoError(t, err)

	version, _, err = migrates.Version()
	assert.NoError(t, err)
	assert.Equal(t, uint64(24), version)

	var migratedUser *models.User
	err = db.Client().Database("test").Collection("users").FindOne(context.TODO(), bson.M{"name": "name"}).Decode(&migratedUser)
	assert.NoError(t, err)
	assert.Equal(t, "username", migratedUser.Username)
	assert.Equal(t, "email@mail.com", migratedUser.Email)

	err = db.Client().Database("test").Collection("users").FindOne(context.TODO(), bson.M{"name": "name2"}).Decode(&migratedUser)
	assert.NoError(t, err)
	assert.Equal(t, "username2", migratedUser.Username)
	assert.Equal(t, "email@mail-test.com", migratedUser.Email)

	err = db.Client().Database("test").Collection("users").FindOne(context.TODO(), bson.M{"name": "name3"}).Decode(&migratedUser)
	assert.NoError(t, err)
	assert.Equal(t, "username3", migratedUser.Username)
	assert.Equal(t, "email@e-mail.com", migratedUser.Email)

	var migratedNamespace *models.Namespace
	err = db.Client().Database("test").Collection("namespaces").FindOne(context.TODO(), bson.M{"tenant_id": "tenant"}).Decode(&migratedNamespace)
	assert.NoError(t, err)
	assert.Equal(t, "name", migratedNamespace.Name)

	err = db.Client().Database("test").Collection("namespaces").FindOne(context.TODO(), bson.M{"tenant_id": "tenant2"}).Decode(&migratedNamespace)
	assert.NoError(t, err)
	assert.Equal(t, "test", migratedNamespace.Name)

	err = db.Client().Database("test").Collection("namespaces").FindOne(context.TODO(), bson.M{"tenant_id": "tenant3"}).Decode(&migratedNamespace)
	assert.NoError(t, err)
	assert.Equal(t, "teste", migratedNamespace.Name)

	err = db.Client().Database("test").Collection("namespaces").FindOne(context.TODO(), bson.M{"username": "USERNAME"}).Decode(&models.Namespace{})
	assert.EqualError(t, mongo.ErrNoDocuments, err.Error())
}
