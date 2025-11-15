package mongo

import (
	"context"
	"io"
	"slices"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func CountAllMatchingDocuments(ctx context.Context, collection *mongo.Collection, basePipeline []bson.M) (int, error) {
	excludeStages := []string{"$skip", "$limit", "$sort"}
	countPipeline := make([]bson.M, 0)

	for _, stage := range basePipeline {
		filtered := make(bson.M)
		for key, value := range stage {
			if !slices.Contains(excludeStages, key) {
				filtered[key] = value
			}
		}

		if len(filtered) > 0 {
			countPipeline = append(countPipeline, filtered)
		}
	}

	countPipeline = append(countPipeline, bson.M{"$count": "count"})
	cursor, err := collection.Aggregate(ctx, countPipeline)
	if err != nil {
		return 0, err
	}
	defer cursor.Close(ctx)

	if !cursor.Next(ctx) {
		return 0, nil
	}

	result := make(map[string]any)
	if err = cursor.Decode(&result); err != nil {
		return 0, err
	}

	return int(result["count"].(int32)), nil
}

// ErrLayer is an error level. Each error defined at this level, is container to it.
// ErrLayer is the errors' level for mongo's error.
const ErrLayer = "mongo"

// ErrMongo is the error for any unknown mongo error.
var ErrMongo = errors.New("mongo error", ErrLayer, 1)

func FromMongoError(err error) error {
	switch {
	case err == mongo.ErrNoDocuments, err == io.EOF:
		return store.ErrNoDocuments
	case err == primitive.ErrInvalidHex:
		return store.ErrInvalidHex
	case mongo.IsDuplicateKeyError(err):
		return store.ErrDuplicate
	default:
		if err == nil {
			return nil
		}

		return errors.Wrap(ErrMongo, err)
	}
}
