package mongo

import (
	"context"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
)

func (s *Store) AnnouncementCreate(ctx context.Context, announcement *models.Announcement) error {
	if _, err := s.db.Collection("announcements").InsertOne(ctx, announcement); err != nil {
		return FromMongoError(err)
	}

	return nil
}

func (s *Store) AnnouncementUpdate(ctx context.Context, announcement *models.Announcement) error {
	result, err := s.db.Collection("announcements").UpdateOne(ctx, bson.M{"uuid": announcement.UUID}, bson.M{"$set": bson.M{"title": announcement.Title, "content": announcement.Content}})
	if err != nil {
		return FromMongoError(err)
	}

	if result.MatchedCount < 1 {
		return store.ErrNoDocuments
	}

	return nil
}

func (s *Store) AnnouncementDelete(ctx context.Context, uuid string) error {
	result, err := s.db.Collection("announcements").DeleteOne(ctx, bson.M{"uuid": uuid})
	if err != nil {
		return FromMongoError(err)
	}

	if result.DeletedCount < 1 {
		return store.ErrNoDocuments
	}

	return nil
}
