package mongo

import (
	"context"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (s *Store) PublicKeyResolve(ctx context.Context, resolver store.PublicKeyResolver, value string, opts ...store.QueryOption) (*models.PublicKey, error) {
	var fingerprint string
	switch resolver {
	case store.PublicKeyFingerprintResolver:
		fingerprint = value
	default:
		return nil, store.ErrNoDocuments
	}

	pipeline := []bson.M{
		{
			"$match": bson.M{
				"fingerprint": fingerprint,
			},
		},
		{
			"$lookup": bson.M{
				"from":         "tags",
				"localField":   "filter.tag_ids",
				"foreignField": "_id",
				"as":           "filter.tags",
			},
		},
	}

	for _, opt := range opts {
		if err := opt(context.WithValue(ctx, "query", &pipeline)); err != nil {
			return nil, err
		}
	}

	cursor, err := s.db.Collection("public_keys").Aggregate(ctx, pipeline)
	if err != nil {
		return nil, FromMongoError(err)
	}
	defer cursor.Close(ctx)

	if !cursor.Next(ctx) {
		return nil, store.ErrNoDocuments
	}

	pubKey := new(models.PublicKey)
	if err := cursor.Decode(&pubKey); err != nil {
		return nil, FromMongoError(err)
	}

	return pubKey, nil
}

func (s *Store) PublicKeyList(ctx context.Context, opts ...store.QueryOption) ([]models.PublicKey, int, error) {
	query := []bson.M{
		{
			"$lookup": bson.M{
				"from":         "tags",
				"localField":   "filter.tag_ids",
				"foreignField": "_id",
				"as":           "filter.tags",
			},
		},
	}

	for _, opt := range opts {
		if err := opt(context.WithValue(ctx, "query", &query)); err != nil {
			return nil, 0, err
		}
	}

	count, err := CountAllMatchingDocuments(ctx, s.db.Collection("public_keys"), query)
	if err != nil {
		return nil, 0, err
	}

	list := make([]models.PublicKey, 0)
	cursor, err := s.db.Collection("public_keys").Aggregate(ctx, query)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		key := new(models.PublicKey)
		err = cursor.Decode(&key)
		if err != nil {
			return list, count, err
		}

		list = append(list, *key)
	}

	return list, count, err
}

func (s *Store) PublicKeyCreate(ctx context.Context, key *models.PublicKey) (string, error) {
	bsonBytes, err := bson.Marshal(key)
	if err != nil {
		return "", FromMongoError(err)
	}

	doc := make(bson.M)
	if err := bson.Unmarshal(bsonBytes, &doc); err != nil {
		return "", FromMongoError(err)
	}

	// WORKAROUND: Convert string TagIDs to MongoDB ObjectIDs for referential integrity
	// with the tags collection where _id is ObjectID type
	if len(key.Filter.TagIDs) > 0 {
		tagIDs := doc["filter"].(bson.M)["tag_ids"].(bson.A)
		for i, id := range tagIDs {
			objID, _ := primitive.ObjectIDFromHex(id.(string))
			tagIDs[i] = objID
		}
	}

	if _, err := s.db.Collection("public_keys").InsertOne(ctx, doc); err != nil {
		return "", FromMongoError(err)
	}

	return doc["fingerprint"].(string), nil
}

func (s *Store) PublicKeyUpdate(ctx context.Context, publicKey *models.PublicKey) error {
	bsonBytes, err := bson.Marshal(publicKey)
	if err != nil {
		return FromMongoError(err)
	}

	doc := make(bson.M)
	if err := bson.Unmarshal(bsonBytes, &doc); err != nil {
		return FromMongoError(err)
	}

	delete(doc, "_id")
	// WORKAROUND: Convert string TagIDs to MongoDB ObjectIDs for referential integrity
	// with the tags collection where _id is ObjectID type
	delete(doc, "tags")
	if filterDoc, ok := doc["filter"].(bson.M); ok {
		if tagIDs, ok := filterDoc["tag_ids"].(bson.A); ok && len(tagIDs) > 0 {
			for i, id := range tagIDs {
				if idStr, ok := id.(string); ok {
					objID, _ := primitive.ObjectIDFromHex(idStr)
					tagIDs[i] = objID
				}
			}
		}
	}

	filter := bson.M{"fingerprint": publicKey.Fingerprint, "tenant_id": publicKey.TenantID}
	r, err := s.db.Collection("public_keys").UpdateOne(ctx, filter, bson.M{"$set": doc})
	if err != nil {
		return FromMongoError(err)
	}

	if r.MatchedCount == 0 {
		return store.ErrNoDocuments
	}

	return nil
}

func (s *Store) PublicKeyDelete(ctx context.Context, publicKey *models.PublicKey) error {
	r, err := s.db.Collection("public_keys").DeleteOne(ctx, bson.M{"fingerprint": publicKey.Fingerprint, "tenant_id": publicKey.TenantID})
	if err != nil {
		return FromMongoError(err)
	}

	if r.DeletedCount < 1 {
		return store.ErrNoDocuments
	}

	return nil
}
