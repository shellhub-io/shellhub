package mongo

import (
	"context"
	"time"

	"github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/models"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	SystemCollection = "system"
	SystemCacheTTL   = 60 * time.Minute
)

func (s *Store) SystemGet(ctx context.Context) (*models.System, error) {
	if system, err := cache.Get[models.System](ctx, s.cache, SystemCollection); err == nil {
		log.WithField("system", system).Warn("using system from cache")

		return system, nil
	}

	result := s.db.Collection(SystemCollection).FindOne(ctx, bson.M{})
	if result.Err() != nil {
		return nil, FromMongoError(result.Err())
	}

	var system *models.System
	if err := result.Decode(&system); err != nil {
		return nil, FromMongoError(err)
	}

	if err := s.cache.Set(ctx, SystemCollection, system, SystemCacheTTL); err != nil {
		log.WithField("system", system).Warn("failed to set the system data on cache")
	}

	return system, nil
}

func (s *Store) SystemSet(ctx context.Context, system *models.System) error {
	upsert := true
	_, err := s.db.Collection(SystemCollection).UpdateOne(ctx, bson.M{}, bson.M{"$set": system}, &options.UpdateOptions{Upsert: &upsert})
	if err != nil {
		return FromMongoError(err)
	}

	if err := s.cache.Delete(ctx, SystemCollection); err != nil {
		log.WithField("system", system).Warn("failed to delete system from cache")
	}

	return nil
}
