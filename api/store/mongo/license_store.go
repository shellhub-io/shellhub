package mongo

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (s *Store) LicenseLoad(ctx context.Context) (*models.License, error) {
	findOpts := options.FindOne()
	findOpts.SetSort(bson.M{"created_at": -1})

	license := new(models.License)
	if err := s.db.Collection("licenses").FindOne(ctx, bson.M{}, findOpts).Decode(&license); err != nil {
		return nil, fromMongoError(err)
	}

	return license, nil
}

func (s *Store) LicenseSave(ctx context.Context, license *models.License) error {
	_, err := s.db.Collection("licenses").InsertOne(ctx, license)
	return fromMongoError(err)
}
