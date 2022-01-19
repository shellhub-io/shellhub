package mongo

import (
	"context"

	"github.com/shellhub-io/shellhub/api/store"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// aggregateCount takes a pipeline and count the results.
func aggregateCount(ctx context.Context, coll *mongo.Collection, pipeline []bson.M) (int, error) {
	resp := struct {
		Count int `bson:"count"`
	}{}

	cursor, err := coll.Aggregate(ctx, pipeline)
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

func fromMongoError(err error) error {
	switch {
	case err == mongo.ErrNoDocuments:
		return store.ErrNoDocuments
	case err == primitive.ErrInvalidHex:
		return store.ErrInvalidHex
	case mongo.IsDuplicateKeyError(err):
		return store.ErrDuplicate
	default:
		return err
	}
}
