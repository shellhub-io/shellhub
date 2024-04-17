package migrations

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
)

func TestMigration63(t *testing.T) {
	t.Cleanup(func() {
		assert.NoError(t, srv.Reset())
	})

	user := models.User{
		UserData: models.UserData{
			Name: "Test",
		},
	}

	_, err := c.Database("test").Collection("users").InsertOne(context.TODO(), user)
	assert.NoError(t, err)

	migrations := GenerateMigrations()[62:63]

	migrates := migrate.NewMigrate(c.Database("test"), migrations...)
	err = migrates.Up(context.Background(), migrate.AllAvailable)
	assert.NoError(t, err)

	version, _, err := migrates.Version(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, uint64(63), version)

	var migratedUser *models.User
	err = c.Database("test").Collection("users").FindOne(context.TODO(), bson.M{"name": user.Name}).Decode(&migratedUser)
	assert.NoError(t, err)
	assert.False(t, migratedUser.MFA)
	assert.Equal(t, "", migratedUser.Secret)
	assert.Empty(t, migratedUser.Codes)

	err = migrates.Down(context.Background(), migrate.AllAvailable)
	assert.NoError(t, err)

	err = c.Database("test").Collection("users").FindOne(context.TODO(), bson.M{"name": user.Name}).Decode(&migratedUser)
	assert.NoError(t, err)
	assert.False(t, migratedUser.MFA)
	assert.Equal(t, "", migratedUser.Secret)
	assert.Empty(t, migratedUser.Codes)
}
