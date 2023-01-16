package stores

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
)

type MongoStore struct {
	Database *mongo.Database
}

func NewMongoStore(ctx context.Context, uri string) (*MongoStore, error) {
	mongoString, err := connstring.ParseAndValidate(uri)
	if err != nil {
		return nil, err
	}

	mongoClientOptions := options.Client().ApplyURI(uri)

	mongoClient, err := mongo.Connect(ctx, mongoClientOptions)
	if err != nil {
		return nil, err
	}

	if err = mongoClient.Ping(ctx, nil); err != nil {
		return nil, err
	}

	database := mongoClient.Database(mongoString.Database)

	return &MongoStore{
		Database: database,
	}, nil
}
