package mongo

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
)

func (s *Store) PrivateKeyCreate(ctx context.Context, key *models.PrivateKey) error {
	_, err := s.db.Collection("private_keys").InsertOne(ctx, key)

	return fromMongoError(err)
}

func (s *Store) PrivateKeyGet(ctx context.Context, fingerprint string) (*models.PrivateKey, error) {
	privKey := new(models.PrivateKey)
	if err := s.db.Collection("private_keys").FindOne(ctx, bson.M{"fingerprint": fingerprint}).Decode(&privKey); err != nil {
		return nil, fromMongoError(err)
	}

	return privKey, nil
}
