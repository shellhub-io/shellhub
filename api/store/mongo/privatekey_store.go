package mongo

import (
	"context"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func (s *Store) CreatePrivateKey(ctx context.Context, key *models.PrivateKey) error {
	_, err := s.db.Collection("private_keys").InsertOne(ctx, key)
	return err
}

func (s *Store) GetPrivateKey(ctx context.Context, fingerprint string) (*models.PrivateKey, error) {
	privKey := new(models.PrivateKey)
	if err := s.db.Collection("private_keys").FindOne(ctx, bson.M{"fingerprint": fingerprint}).Decode(&privKey); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, store.ErrRecordNotFound
		}

		return nil, err
	}

	return privKey, nil
}
