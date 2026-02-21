package mongo

import (
	"context"
	"errors"
	"fmt"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (s *Store) TagCreate(ctx context.Context, tag *models.Tag) (string, error) {
	id := primitive.NewObjectID()
	upsert := bson.M{
		"$setOnInsert": bson.M{"_id": id},
		"$set": bson.M{
			"name":       tag.Name,
			"tenant_id":  tag.TenantID,
			"created_at": clock.Now(),
			"updated_at": clock.Now(),
		},
	}

	_, err := s.db.
		Collection("tags").
		UpdateOne(ctx, bson.M{"tenant_id": tag.TenantID, "name": tag.Name}, upsert, options.Update().SetUpsert(true))
	if err != nil {
		return "", FromMongoError(err)
	}

	return id.Hex(), nil
}

func (s *Store) TagConflicts(ctx context.Context, tenantID string, target *models.TagConflicts) ([]string, bool, error) {
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"tenant_id": tenantID,
				"$or":       []bson.M{{"name": target.Name}},
			},
		},
	}

	cursor, err := s.db.Collection("tags").Aggregate(ctx, pipeline)
	if err != nil {
		return nil, false, FromMongoError(err)
	}
	defer cursor.Close(ctx)

	tag := new(models.Tag)
	conflicts := make([]string, 0)

	for cursor.Next(ctx) {
		if err := cursor.Decode(&tag); err != nil {
			return nil, false, FromMongoError(err)
		}

		if tag.Name == target.Name {
			conflicts = append(conflicts, "name")
		}
	}

	return conflicts, len(conflicts) > 0, nil
}

func (s *Store) TagList(ctx context.Context, opts ...store.QueryOption) ([]models.Tag, int, error) {
	query := []bson.M{}
	for _, opt := range opts {
		if err := opt(context.WithValue(ctx, "query", &query)); err != nil {
			return nil, 0, err
		}
	}

	count, err := CountAllMatchingDocuments(ctx, s.db.Collection("tags"), query)
	if err != nil {
		return nil, 0, FromMongoError(err)
	}

	if count == 0 {
		return []models.Tag{}, 0, nil
	}

	cursor, err := s.db.Collection("tags").Aggregate(ctx, query)
	if err != nil {
		return []models.Tag{}, 0, err
	}
	defer cursor.Close(ctx)

	tags := make([]models.Tag, 0)
	for cursor.Next(ctx) {
		tag := new(models.Tag)
		if err := cursor.Decode(tag); err != nil {
			return []models.Tag{}, 0, err
		}

		tags = append(tags, *tag)
	}

	return tags, count, err
}

func (s *Store) TagResolve(ctx context.Context, resolver store.TagResolver, value string, opts ...store.QueryOption) (*models.Tag, error) {
	query := []bson.M{}
	switch resolver {
	case store.TagIDResolver:
		objID, err := primitive.ObjectIDFromHex(value)
		if err != nil {
			// Invalid ObjectID format will never exist in database
			// Wrap as ErrNoDocuments while preserving original error details
			return nil, fmt.Errorf("%w: %v", store.ErrNoDocuments, err)
		}

		query = append(query, bson.M{"$match": bson.M{"_id": objID}})
	case store.TagNameResolver:
		query = append(query, bson.M{"$match": bson.M{"name": value}})
	}

	for _, opt := range opts {
		if err := opt(context.WithValue(ctx, "query", &query)); err != nil {
			return nil, err
		}
	}

	cursor, err := s.db.Collection("tags").Aggregate(ctx, query)
	if err != nil {
		return nil, FromMongoError(err)
	}

	defer cursor.Close(ctx)
	cursor.Next(ctx)

	tag := new(models.Tag)
	if err := cursor.Decode(&tag); err != nil {
		return nil, FromMongoError(err)
	}

	return tag, nil
}

func (s *Store) TagUpdate(ctx context.Context, tag *models.Tag) error {
	bsonBytes, err := bson.Marshal(tag)
	if err != nil {
		return FromMongoError(err)
	}

	doc := make(bson.M)
	if err := bson.Unmarshal(bsonBytes, &doc); err != nil {
		return FromMongoError(err)
	}

	objID, _ := primitive.ObjectIDFromHex(tag.ID)
	doc["_id"] = objID

	filter := bson.M{"_id": objID}
	r, err := s.db.Collection("tags").UpdateOne(ctx, filter, bson.M{"$set": doc})
	if err != nil {
		return FromMongoError(err)
	}

	if r.MatchedCount < 1 {
		return store.ErrNoDocuments
	}

	return nil
}

func (s *Store) TagPushToTarget(ctx context.Context, id string, target store.TagTarget, targetID string) error {
	tag, err := s.TagResolve(ctx, store.TagIDResolver, id)
	if err != nil {
		return err
	}

	collection, filter, attribute, err := getTargetCollectionMetadata(target)
	if err != nil {
		return err
	}

	tagID, _ := primitive.ObjectIDFromHex(tag.ID)
	res, err := s.db.
		Collection(collection).
		UpdateOne(ctx, bson.M{filter: targetID}, bson.M{"$addToSet": bson.M{attribute: tagID}})
	if err != nil {
		return FromMongoError(err)
	}

	if res.MatchedCount < 1 {
		return store.ErrNoDocuments
	}

	return nil
}

func (s *Store) TagPullFromTarget(ctx context.Context, id string, target store.TagTarget, targetIDs ...string) error {
	tag, err := s.TagResolve(ctx, store.TagIDResolver, id)
	if err != nil {
		return err
	}

	collection, filter, attribute, err := getTargetCollectionMetadata(target)
	if err != nil {
		return err
	}

	tagID, _ := primitive.ObjectIDFromHex(tag.ID)
	if len(targetIDs) > 0 {
		res, err := s.db.
			Collection(collection).
			UpdateMany(ctx, bson.M{"tenant_id": tag.TenantID, filter: bson.M{"$in": targetIDs}}, bson.M{"$pull": bson.M{attribute: tagID}})
		if err != nil {
			return FromMongoError(err)
		}

		if res.MatchedCount < 1 {
			return store.ErrNoDocuments
		}

		return nil
	} else { // nolint:revive
		_, err = s.db.Collection(collection).UpdateMany(ctx, bson.M{"tenant_id": tag.TenantID}, bson.M{"$pull": bson.M{"tags": tagID}})

		return FromMongoError(err)
	}
}

func (s *Store) TagDelete(ctx context.Context, tag *models.Tag) error {
	session, err := s.db.Client().StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	sessionCallback := func(sessCtx mongo.SessionContext) (any, error) {
		objID, err := primitive.ObjectIDFromHex(tag.ID)
		if err != nil {
			// Invalid ObjectID format will never exist in database
			// Wrap as ErrNoDocuments while preserving original error details
			return nil, fmt.Errorf("%w: %v", store.ErrNoDocuments, err)
		}

		r, err := s.db.Collection("tags").DeleteOne(sessCtx, bson.M{"_id": objID})
		if err != nil {
			return nil, FromMongoError(err)
		}

		if r.DeletedCount < 1 {
			return nil, store.ErrNoDocuments
		}

		if _, err := s.db.Collection("devices").UpdateMany(sessCtx, bson.M{"tenant_id": tag.TenantID}, bson.M{"$pull": bson.M{"tag_ids": objID}}); err != nil {
			return nil, FromMongoError(err)
		}

		if _, err := s.db.Collection("public_keys").UpdateMany(sessCtx, bson.M{"tenant_id": tag.TenantID}, bson.M{"$pull": bson.M{"filters.tag_ids": objID}}); err != nil {
			return nil, FromMongoError(err)
		}

		return nil, nil
	}

	_, err = session.WithTransaction(ctx, sessionCallback)

	return err
}

func getTargetCollectionMetadata(target store.TagTarget) (string, string, string, error) {
	switch target {
	case store.TagTargetDevice:
		return "devices", "uid", "tag_ids", nil
	case store.TagTargetPublicKey:
		return "public_keys", "fingerprint", "filter.tag_ids", nil
	default:
		return "", "", "", errors.New("invalid tag target")
	}
}
