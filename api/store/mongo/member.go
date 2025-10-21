package mongo

import (
	"context"
	"strings"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/models"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func (s *Store) NamespaceCreateMembership(ctx context.Context, tenantID string, member *models.Member) error {
	err := s.db.
		Collection("namespaces").
		FindOne(ctx, bson.M{"tenant_id": tenantID, "members": bson.M{"$elemMatch": bson.M{"id": member.ID}}}).
		Err()
	if err == nil {
		return ErrNamespaceDuplicatedMember
	}

	memberBson := bson.M{
		"id":         member.ID,
		"added_at":   member.AddedAt,
		"expires_at": member.ExpiresAt,
		"role":       member.Role,
		"status":     member.Status,
	}

	res, err := s.db.
		Collection("namespaces").
		UpdateOne(ctx, bson.M{"tenant_id": tenantID}, bson.M{"$addToSet": bson.M{"members": memberBson}})
	if err != nil {
		return FromMongoError(err)
	}

	if res.MatchedCount < 1 {
		return store.ErrNoDocuments
	}

	if err := s.cache.Delete(ctx, strings.Join([]string{"namespace", tenantID}, "/")); err != nil {
		log.Error(err)
	}

	return nil
}

func (s *Store) NamespaceUpdateMembership(ctx context.Context, tenantID string, member *models.Member) error {
	filter := bson.M{"tenant_id": tenantID, "members": bson.M{"$elemMatch": bson.M{"id": member.ID}}}

	memberBson := bson.M{
		"members.$.id":         member.ID,
		"members.$.added_at":   member.AddedAt,
		"members.$.expires_at": member.ExpiresAt,
		"members.$.role":       member.Role,
		"members.$.status":     member.Status,
	}

	ns, err := s.db.Collection("namespaces").UpdateOne(ctx, filter, bson.M{"$set": memberBson})
	if err != nil {
		return FromMongoError(err)
	}

	if ns.MatchedCount < 1 {
		return ErrUserNotFound
	}

	if err := s.cache.Delete(ctx, strings.Join([]string{"namespace", tenantID}, "/")); err != nil {
		log.Error(err)
	}

	return nil
}

func (s *Store) NamespaceDeleteMembership(ctx context.Context, tenantID string, member *models.Member) error {
	session, err := s.db.Client().StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	fn := func(_ mongo.SessionContext) (any, error) {
		res, err := s.db.
			Collection("namespaces").
			UpdateOne(ctx, bson.M{"tenant_id": tenantID}, bson.M{"$pull": bson.M{"members": bson.M{"id": member.ID}}})
		if err != nil {
			return nil, FromMongoError(err)
		}

		switch {
		case res.MatchedCount < 1: // tenant not found
			return nil, store.ErrNoDocuments
		case res.ModifiedCount < 1: // member not found
			return nil, ErrUserNotFound
		}

		objID, err := primitive.ObjectIDFromHex(member.ID)
		if err != nil {
			return nil, err
		}

		_, err = s.db.
			Collection("users").
			UpdateOne(ctx, bson.M{"_id": objID, "preferred_namespace": tenantID}, bson.M{"$set": bson.M{"preferred_namespace": ""}})

		return nil, FromMongoError(err)
	}

	if _, err := session.WithTransaction(ctx, fn); err != nil {
		return err
	}

	if err := s.cache.Delete(ctx, strings.Join([]string{"namespace", tenantID}, "/")); err != nil {
		log.Error(err)
	}

	return nil
}
