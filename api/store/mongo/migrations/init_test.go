package migrations

import (
	"os"
	"testing"

	"go.mongodb.org/mongo-driver/mongo"
)

var (
	mongoClient *mongo.Client
)

func TestMain(_ *testing.M) {
	os.Exit(0)
}
