package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func aggregateCount(ctx context.Context, coll *mongo.Collection, query []bson.M) (int, error) {
	resp := struct {
		Count int `bson:"count"`
	}{}

	cursor, err := coll.Aggregate(ctx, query)
	if err != nil {
		return 0, err
	}

	defer cursor.Close(ctx)

	if !cursor.Next(ctx) {
		return 0, nil
	}

	if err = cursor.Decode(&resp); err != nil {
		return 0, err
	}

	return resp.Count, nil
}
