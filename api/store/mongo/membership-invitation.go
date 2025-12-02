package mongo

import (
	"context"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (s *Store) MembershipInvitationCreate(ctx context.Context, invitation *models.MembershipInvitation) error {
	now := clock.Now()
	invitation.CreatedAt = now
	invitation.UpdatedAt = now
	invitation.StatusUpdatedAt = now

	bsonBytes, err := bson.Marshal(invitation)
	if err != nil {
		return FromMongoError(err)
	}

	doc := make(bson.M)
	if err := bson.Unmarshal(bsonBytes, &doc); err != nil {
		return FromMongoError(err)
	}

	objID := primitive.NewObjectID()
	doc["_id"] = objID
	doc["user_id"], _ = primitive.ObjectIDFromHex(invitation.UserID)
	doc["invited_by"], _ = primitive.ObjectIDFromHex(invitation.InvitedBy)

	if _, err := s.db.Collection("membership_invitations").InsertOne(ctx, doc); err != nil {
		return FromMongoError(err)
	}

	invitation.ID = objID.Hex()

	return nil
}

func (s *Store) MembershipInvitationResolve(ctx context.Context, tenantID, userID string) (*models.MembershipInvitation, error) {
	userObjID, _ := primitive.ObjectIDFromHex(userID)

	pipeline := []bson.M{
		{
			"$match": bson.M{"tenant_id": tenantID, "user_id": userObjID},
		},
		{
			"$sort": bson.D{{Key: "_id", Value: -1}},
		},
		{
			"$limit": 1,
		},
		{
			"$lookup": bson.M{
				"from":         "namespaces",
				"localField":   "tenant_id",
				"foreignField": "tenant_id",
				"as":           "namespace",
			},
		},
		{
			"$lookup": bson.M{
				"from":         "users",
				"localField":   "user_id",
				"foreignField": "_id",
				"as":           "user",
			},
		},
		{
			"$lookup": bson.M{
				"from":         "user_invitations",
				"localField":   "user_id",
				"foreignField": "_id",
				"as":           "user_invitation",
			},
		},
		{
			"$addFields": bson.M{
				"namespace_name": bson.M{"$arrayElemAt": bson.A{"$namespace.name", 0}},
				"user_email": bson.M{
					"$ifNull": bson.A{
						bson.M{"$arrayElemAt": bson.A{"$user.email", 0}},
						bson.M{"$arrayElemAt": bson.A{"$user_invitation.email", 0}},
					},
				},
			},
		},
		{
			"$project": bson.M{
				"namespace":       0,
				"user":            0,
				"user_invitation": 0,
			},
		},
	}

	cursor, err := s.db.Collection("membership_invitations").Aggregate(ctx, pipeline)
	if err != nil {
		return nil, FromMongoError(err)
	}
	defer cursor.Close(ctx)

	if !cursor.Next(ctx) {
		return nil, store.ErrNoDocuments
	}

	invitation := &models.MembershipInvitation{}
	if err := cursor.Decode(invitation); err != nil {
		return nil, FromMongoError(err)
	}

	return invitation, nil
}

func (s *Store) MembershipInvitationUpdate(ctx context.Context, invitation *models.MembershipInvitation) error {
	invitation.UpdatedAt = clock.Now()

	bsonBytes, err := bson.Marshal(invitation)
	if err != nil {
		return FromMongoError(err)
	}

	doc := make(bson.M)
	if err := bson.Unmarshal(bsonBytes, &doc); err != nil {
		return FromMongoError(err)
	}

	delete(doc, "_id")
	doc["user_id"], _ = primitive.ObjectIDFromHex(invitation.UserID)
	doc["invited_by"], _ = primitive.ObjectIDFromHex(invitation.InvitedBy)

	objID, _ := primitive.ObjectIDFromHex(invitation.ID)
	r, err := s.db.Collection("membership_invitations").UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": doc})
	if err != nil {
		return FromMongoError(err)
	}

	if r.MatchedCount == 0 {
		return store.ErrNoDocuments
	}

	return nil
}
