package migrations

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
)

func TestMigration27(t *testing.T) {
	t.Cleanup(func() {
		assert.NoError(t, srv.Reset())
	})

	migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[:26]...)
	err := migrates.Up(context.Background(), migrate.AllAvailable)
	assert.NoError(t, err)

	sessionsToBeMigrated := []struct {
		UID string `json:"uid"`
	}{
		{
			UID: "uid1",
		},
		{
			UID: "uid2",
		},
		{
			UID: "uid3",
		},
	}

	sessions := make([]interface{}, len(sessionsToBeMigrated))
	for i, v := range sessionsToBeMigrated {
		sessions[i] = v
	}

	_, err = c.Database("test").Collection("sessions").InsertMany(context.TODO(), sessions)
	assert.NoError(t, err)

	migrates = migrate.NewMigrate(c.Database("test"), GenerateMigrations()[:27]...)
	err = migrates.Up(context.Background(), migrate.AllAvailable)
	assert.NoError(t, err)

	migratedSessions := []models.Session{}
	cur, err := c.Database("test").Collection("sessions").Find(context.TODO(), bson.D{})
	assert.NoError(t, err)
	for cur.Next(context.TODO()) {
		var ses models.Session
		err := cur.Decode(&ses)
		if err != nil {
			panic(err.Error())
		}
		migratedSessions = append(migratedSessions, ses)
	}

	for _, ses := range migratedSessions {
		assert.Equal(t, true, ses.Closed)
	}
}
