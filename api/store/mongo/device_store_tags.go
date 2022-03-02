package mongo

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
)

func (s *Store) DeviceCreateTag(ctx context.Context, uid models.UID, tag string) error {
	_, err := s.db.Collection("devices").UpdateOne(ctx, bson.M{"uid": uid}, bson.M{"$push": bson.M{"tags": tag}})

	return err
}

func (s *Store) DeviceRemoveTag(ctx context.Context, uid models.UID, tag string) error {
	_, err := s.db.Collection("devices").UpdateOne(ctx, bson.M{"uid": uid}, bson.M{"$pull": bson.M{"tags": tag}})

	return err
}

func (s *Store) DeviceUpdateTag(ctx context.Context, uid models.UID, tags []string) error {
	_, err := s.db.Collection("devices").UpdateOne(ctx, bson.M{"uid": uid}, bson.M{"$set": bson.M{"tags": tags}})

	return err
}
