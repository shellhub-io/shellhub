package migrations

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	"github.com/stretchr/testify/assert"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestMigration26(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	migrations := GenerateMigrations()[:26]

	migrates := migrate.NewMigrate(db.Client().Database("test"), migrations...)
	err := migrates.Up(migrate.AllAvailable)
	assert.NoError(t, err)

	version, _, err := migrates.Version()
	assert.NoError(t, err)
	assert.Equal(t, uint64(26), version)

	userToken := models.UserTokenRecover{
		Token:     uuid.Generate(),
		User:      "user",
		CreatedAt: clock.Now(),
	}
	_, err = db.Client().Database("test").Collection("recovery_tokens").InsertOne(context.TODO(), userToken)
	assert.NoError(t, err)

	var migratedUserToken *models.UserTokenRecover
	err = db.Client().Database("test").Collection("recovery_tokens").FindOne(context.TODO(), bson.M{"user": userToken.User}).Decode(&migratedUserToken)
	assert.NoError(t, err)
	assert.Equal(t, userToken.Token, migratedUserToken.Token)

	index := db.Client().Database("test").Collection("recovery_tokens").Indexes()

	cursor, err := index.List(context.TODO())
	assert.NoError(t, err)

	var results []bson.M
	err = cursor.All(context.TODO(), &results)
	assert.NoError(t, err)

	keyField := results[1]["key"].(primitive.M)
	assert.Equal(t, int32(1), keyField["created_at"])

	keyField = results[2]["key"].(primitive.M)
	assert.Equal(t, int32(1), keyField["token"])

	value, key := results[1]["expireAfterSeconds"]
	assert.Equal(t, true, key)
	assert.Equal(t, int32(86400), value)

	value, key = results[1]["name"]
	assert.Equal(t, true, key)
	assert.Equal(t, "ttl", value)

	value, key = results[2]["name"]
	assert.Equal(t, true, key)
	assert.Equal(t, "token", value)

	value, key = results[3]["name"]
	assert.Equal(t, true, key)
	assert.Equal(t, "user", value)
}
