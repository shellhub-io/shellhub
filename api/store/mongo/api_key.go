package mongo

import (
	"context"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/mongo/queries"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func (s *Store) APIKeyCreate(ctx context.Context, req *models.APIKey) error {
	_, err := s.db.Collection("api_keys").InsertOne(ctx, req)

	return FromMongoError(err)
}

func (s *Store) APIKeyList(ctx context.Context, userID string, paginator query.Paginator, sorter query.Sorter, tenantID string) ([]models.APIKey, int, error) {
	query := []bson.M{}

	query = append(query, bson.M{"$match": bson.M{"user_id": userID, "tenant_id": tenantID}})
	queryCount := append(query, bson.M{"$count": "count"})

	count, err := AggregateCount(ctx, s.db.Collection("api_keys"), queryCount)
	if err != nil {
		return nil, 0, FromMongoError(err)
	}

	query = append(query, queries.FromSorter(&sorter)...)
	query = append(query, queries.FromPaginator(&paginator)...)
	cursor, err := s.db.Collection("api_keys").Aggregate(ctx, query)
	if err != nil {
		return nil, 0, FromMongoError(err)
	}
	defer cursor.Close(ctx)

	apiKeys := make([]models.APIKey, 0)
	for cursor.Next(ctx) {
		var apiKey models.APIKey
		if err := cursor.Decode(&apiKey); err != nil {
			return nil, 0, FromMongoError(err)
		}
		apiKeys = append(apiKeys, models.APIKey{
			ID:        apiKey.ID,
			UserID:    apiKey.UserID,
			TenantID:  apiKey.TenantID,
			Name:      apiKey.Name,
			ExpiresIn: apiKey.ExpiresIn,
		})
	}

	if err := cursor.Err(); err != nil {
		return nil, 0, FromMongoError(err)
	}

	return apiKeys, count, nil
}

func (s *Store) APIKeyGetByUID(ctx context.Context, uid string) (*models.APIKey, error) {
	var APIKey *models.APIKey

	if err := s.db.Collection("api_keys").FindOne(ctx, bson.M{"_id": uid}).Decode(&APIKey); err != nil {
		return nil, FromMongoError(err)
	}

	return APIKey, nil
}

func (s *Store) APIKeyGetByName(ctx context.Context, name string) (*models.APIKey, error) {
	var APIKey models.APIKey

	err := s.db.Collection("api_keys").FindOne(ctx, bson.M{"name": name}).Decode(&APIKey)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}

		return nil, FromMongoError(err)
	}

	return &APIKey, nil
}

func (s *Store) APIKeyDelete(ctx context.Context, id string, tenantID string) error {
	result, err := s.db.Collection("api_keys").DeleteOne(ctx, bson.M{"_id": id, "tenant_id": tenantID})
	if err != nil {
		return FromMongoError(err)
	}

	if result.DeletedCount < 1 {
		return store.ErrNoDocuments
	}

	return nil
}

func (s *Store) APIKeyEdit(ctx context.Context, changes *requests.APIKeyChanges) error {
	updatedFields := bson.M{}

	if changes.Name != "" {
		updatedFields["name"] = changes.Name
	}

	if len(updatedFields) > 0 {
		key, err := s.db.Collection("api_keys").UpdateOne(ctx, bson.M{"_id": changes.ID}, bson.M{"$set": updatedFields})
		if err != nil {
			return FromMongoError(err)
		}

		if key.ModifiedCount < 1 {
			return store.ErrNoDocuments
		}
	}

	return nil
}
