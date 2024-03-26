package dbtest

import (
	"context"
	"log"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Container struct {
	Client *mongo.Client
	Host   string
	Stop   func()
}

func Setup(ctx context.Context) (*Container, error) {
	client, host, stop, err := start(ctx)
	if err != nil {
		log.Panicf("failed to start test container with err: %s", err)
		return nil, err
	}

	return &Container{
		Client: client,
		Host:   host,
		Stop:   stop,
	}, nil
}

func start(ctx context.Context) (*mongo.Client, string, func(), error) {
	container, err := mongodb.RunContainer(ctx, testcontainers.WithImage("mongo:4.4.8"))
	if err != nil {
		log.Panicf("failed to start test container MongoDB: %s", err)
	}

	endpoint, err := container.ConnectionString(ctx)
	if err != nil {
		log.Panicf("failed to get connection string for test container MongoDB: %s", err)
	}

	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(endpoint))
	if err != nil {
		log.Panicf("failed to connect to test container MongoDB: %s", err)
	}

	err = mongoClient.Ping(ctx, nil)
	if err != nil {
		log.Panicf("failed to ping to test container MongoDB: %s", err)
	}

	mongoClient.StartSession()
	mongoClient.UseSession(ctx, func(sctx mongo.SessionContext) error {
		err := sctx.StartTransaction()
		if err != nil {
			return err
		}
		return nil
	})

	containerIP, err := container.Host(ctx)
	if err != nil {
		log.Panicf("failed to get test container IP address: %s", err)
	}

	stopContainer := func() {
		if err := container.Terminate(ctx); err != nil {
			log.Panicf("failed to stop test container MongoDB: %s", err)
		}
	}

	return mongoClient, containerIP, stopContainer, err
}

func InsertMockData(ctx context.Context, collection *mongo.Collection, data []interface{}) error {
	_, err := collection.InsertMany(ctx, data)
	return err
}

func DeleteMockData(ctx context.Context, collection *mongo.Collection) error {
	_, err := collection.DeleteMany(ctx, bson.M{})
	return err
}

func FindMockData(ctx context.Context, collection *mongo.Collection, filter interface{}) ([]interface{}, error) {
	cur, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var results []interface{}
	for cur.Next(ctx) {
		var result interface{}
		if err := cur.Decode(&result); err != nil {
			return nil, err
		}
		results = append(results, result)
	}

	if err := cur.Err(); err != nil {
		return nil, err
	}

	return results, nil
}
