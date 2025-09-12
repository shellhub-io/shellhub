package mongo

import (
	"context"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
)

func (s *Store) APIKeyCreate(ctx context.Context, apiKey *models.APIKey) (string, error) {
	now := clock.Now()
	apiKey.CreatedAt = now
	apiKey.UpdatedAt = now

	res, err := s.db.Collection("api_keys").InsertOne(ctx, apiKey)
	if err != nil {
		return "", FromMongoError(err)
	}

	return res.InsertedID.(string), nil
}

func (s *Store) APIKeyConflicts(ctx context.Context, tenantID string, target *models.APIKeyConflicts) ([]string, bool, error) {
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"tenant_id": tenantID,
				"$or": []bson.M{
					{"_id": target.ID},
					{"name": target.Name},
				},
			},
		},
	}

	cursor, err := s.db.Collection("api_keys").Aggregate(ctx, pipeline)
	if err != nil {
		return nil, false, FromMongoError(err)
	}
	defer cursor.Close(ctx)

	apiKey := new(models.APIKeyConflicts)
	conflicts := make([]string, 0)
	for cursor.Next(ctx) {
		if err := cursor.Decode(&apiKey); err != nil {
			return nil, false, FromMongoError(err)
		}

		if apiKey.ID == target.ID {
			conflicts = append(conflicts, "id")
		}

		if apiKey.Name == target.Name {
			conflicts = append(conflicts, "name")
		}
	}

	return conflicts, len(conflicts) > 0, nil
}

func (s *Store) APIKeyResolve(ctx context.Context, resolver store.APIKeyResolver, value string, opts ...store.QueryOption) (*models.APIKey, error) {
	query := []bson.M{}
	switch resolver {
	case store.APIKeyIDResolver:
		query = append(query, bson.M{"$match": bson.M{"_id": value}})
	case store.APIKeyNameResolver:
		query = append(query, bson.M{"$match": bson.M{"name": value}})
	}

	for _, opt := range opts {
		if err := opt(context.WithValue(ctx, "query", &query)); err != nil {
			return nil, err
		}
	}

	cursor, err := s.db.Collection("api_keys").Aggregate(ctx, query)
	if err != nil {
		return nil, FromMongoError(err)
	}
	defer cursor.Close(ctx)

	if !cursor.Next(ctx) {
		return nil, store.ErrNoDocuments
	}

	apiKey := new(models.APIKey)
	if err := cursor.Decode(&apiKey); err != nil {
		return nil, FromMongoError(err)
	}

	return apiKey, nil
}

func (s *Store) APIKeyList(ctx context.Context, opts ...store.QueryOption) ([]models.APIKey, int, error) {
	query := []bson.M{}
	for _, opt := range opts {
		if err := opt(context.WithValue(ctx, "query", &query)); err != nil {
			return nil, 0, err
		}
	}

	count, err := CountAllMatchingDocuments(ctx, s.db.Collection("api_keys"), query)
	if err != nil {
		return nil, 0, FromMongoError(err)
	}

	if count == 0 {
		return []models.APIKey{}, 0, nil
	}

	cursor, err := s.db.Collection("api_keys").Aggregate(ctx, query)
	if err != nil {
		return nil, 0, FromMongoError(err)
	}
	defer cursor.Close(ctx)

	apiKeys := make([]models.APIKey, 0)
	for cursor.Next(ctx) {
		apiKey := new(models.APIKey)
		if err := cursor.Decode(apiKey); err != nil {
			return nil, 0, FromMongoError(err)
		}

		apiKeys = append(apiKeys, *apiKey)
	}

	return apiKeys, count, nil
}

func (s *Store) APIKeyUpdate(ctx context.Context, apiKey *models.APIKey) error {
	apiKey.UpdatedAt = clock.Now()

	res, err := s.db.Collection("api_keys").UpdateOne(ctx, bson.M{"_id": apiKey.ID}, bson.M{"$set": apiKey})
	if err != nil {
		return FromMongoError(err)
	}

	if res.ModifiedCount < 1 {
		return store.ErrNoDocuments
	}

	return nil
}

func (s *Store) APIKeyDelete(ctx context.Context, apiKey *models.APIKey) error {
	res, err := s.db.Collection("api_keys").DeleteOne(ctx, bson.M{"_id": apiKey.ID})
	if err != nil {
		return FromMongoError(err)
	}

	if res.DeletedCount < 1 {
		return store.ErrNoDocuments
	}

	return nil
}
