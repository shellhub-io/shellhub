package mongo

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GetStatusMFA seachr for statusMFA in the lits of users by id.
func (s *Store) GetStatusMFA(ctx context.Context, id string) (bool, error) {
	var user models.User
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return false, err
	}

	if err := s.db.Collection("users").FindOne(ctx, bson.M{"_id": objID}).Decode(&user); err != nil {
		return false, FromMongoError(err)
	}

	return user.MFA, nil
}

// Add a new StatusMFA for the user by email.
func (s *Store) AddStatusMFA(ctx context.Context, username string, statusMFA bool) error {
	if _, err := s.db.Collection("users").UpdateOne(ctx, bson.M{"username": username}, bson.M{"$set": bson.M{"status_mfa": statusMFA}}); err != nil {
		return FromMongoError(err)
	}

	return nil
}

func (s *Store) AddSecret(ctx context.Context, username string, secret string) error {
	if _, err := s.db.Collection("users").UpdateOne(ctx, bson.M{"username": username}, bson.M{"$set": bson.M{"secret": secret}}); err != nil {
		return FromMongoError(err)
	}

	return nil
}

func (s *Store) DeleteSecret(ctx context.Context, username string) error {
	_, err := s.db.Collection("users").UpdateOne(ctx, bson.M{"username": username}, bson.M{"$unset": bson.M{"secret": ""}})
	if err != nil {
		return FromMongoError(err)
	}

	return nil
}
