package mongo

import (
	"context"
	"strings"

	"github.com/shellhub-io/shellhub/pkg/clock"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (s *Store) UserInvitationsUpsert(ctx context.Context, email string) (string, error) {
	now := clock.Now()

	r := s.db.Collection("user_invitations").FindOneAndUpdate(
		ctx,
		bson.M{
			"email": email,
		},
		bson.M{
			"$setOnInsert": bson.M{
				"created_at": now,
				"status":     "pending",
			},
			"$set": bson.M{
				"email":      strings.ToLower(email),
				"updated_at": now,
			},
			"$inc": bson.M{
				"invitations": 1,
			},
		},
		options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After),
	)

	userInvitation := make(bson.M)
	if err := r.Decode(&userInvitation); err != nil {
		return "", FromMongoError(err)
	}

	return userInvitation["_id"].(primitive.ObjectID).Hex(), nil
}
