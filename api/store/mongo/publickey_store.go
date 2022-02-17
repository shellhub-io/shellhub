package mongo

import (
	"context"

	"github.com/shellhub-io/shellhub/api/pkg/gateway"
	"github.com/shellhub-io/shellhub/api/store/mongo/queries"
	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
)

func (s *Store) PublicKeyGet(ctx context.Context, fingerprint string, tenantID string) (*models.PublicKey, error) {
	pubKey := new(models.PublicKey)
	if tenantID != "" {
		if err := s.db.Collection("public_keys").FindOne(ctx, bson.M{"fingerprint": fingerprint, "tenant_id": tenantID}).Decode(&pubKey); err != nil {
			return nil, fromMongoError(err)
		}
	} else {
		if err := s.db.Collection("public_keys").FindOne(ctx, bson.M{"fingerprint": fingerprint}).Decode(&pubKey); err != nil {
			return nil, fromMongoError(err)
		}
	}

	return pubKey, nil
}

func (s *Store) PublicKeyList(ctx context.Context, pagination paginator.Query) ([]models.PublicKey, int, error) {
	query := []bson.M{
		{
			"$sort": bson.M{
				"created_at": 1,
			},
		},
	}

	// Only match for the respective tenant if requested
	if tenant := gateway.TenantFromContext(ctx); tenant != nil {
		query = append(query, bson.M{
			"$match": bson.M{
				"tenant_id": tenant.ID,
			},
		})
	}

	queryCount := query
	queryCount = append(queryCount, bson.M{"$count": "count"})
	count, err := aggregateCount(ctx, s.db.Collection("public_keys"), queryCount)
	if err != nil {
		return nil, 0, err
	}

	query = append(query, queries.BuildPaginationQuery(pagination)...)

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

func (s *Store) PublicKeyCreate(ctx context.Context, key *models.PublicKey) error {
	_, err := s.db.Collection("public_keys").InsertOne(ctx, key)

	return fromMongoError(err)
}

func (s *Store) PublicKeyUpdate(ctx context.Context, fingerprint string, tenantID string, key *models.PublicKeyUpdate) (*models.PublicKey, error) {
	if _, err := s.db.Collection("public_keys").UpdateOne(ctx, bson.M{"fingerprint": fingerprint}, bson.M{"$set": key}); err != nil {
		if err != nil {
			return nil, fromMongoError(err)
		}

		return nil, err
	}

	return s.PublicKeyGet(ctx, fingerprint, tenantID)
}

func (s *Store) PublicKeyDelete(ctx context.Context, fingerprint string, tenantID string) error {
	_, err := s.db.Collection("public_keys").DeleteOne(ctx, bson.M{"fingerprint": fingerprint, "tenant_id": tenantID})

	return err
}
