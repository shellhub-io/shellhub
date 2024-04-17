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

func TestMigration35(t *testing.T) {
	t.Cleanup(func() {
		assert.NoError(t, fixtures.Teardown())
	})

	type User struct {
		ID            string          `json:"id,omitempty" bson:"_id,omitempty"`
		Namespaces    int             `json:"namespaces" bson:"namespaces,omitempty"`
		Authenticated bool            `json:"authenticated"`
		UserData      models.UserData `bson:",inline"`
	}

	user := User{
		ID:            "0",
		Namespaces:    0,
		Authenticated: true,
		UserData: models.UserData{
			Name:     "user",
			Email:    "test@shellhub.com",
			Username: "username",
		},
	}

	_, err := srv.Client().Database("test").Collection("users").InsertOne(context.TODO(), user)
	assert.NoError(t, err)

	var afterMigrationUser *User
	err = srv.Client().Database("test").Collection("users").FindOne(context.TODO(), bson.M{"username": "username"}).Decode(&afterMigrationUser)
	assert.NoError(t, err)

	migrates := migrate.NewMigrate(srv.Client().Database("test"), GenerateMigrations()[34:35]...)
	err = migrates.Up(context.Background(), migrate.AllAvailable)
	assert.NoError(t, err)

	var migratedUser *models.User
	err = srv.Client().Database("test").Collection("users").FindOne(context.TODO(), bson.M{"username": "username"}).Decode(&migratedUser)
	assert.NoError(t, err)
}
