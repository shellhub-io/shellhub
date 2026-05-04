package mongo

import (
	"context"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
)

func (s *Store) OAuthClientCreate(ctx context.Context, client *models.OAuthClient) (string, error) {
	now := clock.Now()
	client.CreatedAt = now
	client.UpdatedAt = now

	res, err := s.db.Collection("oauth_clients").InsertOne(ctx, client)
	if err != nil {
		return "", FromMongoError(err)
	}

	return res.InsertedID.(string), nil
}

func (s *Store) OAuthClientResolve(ctx context.Context, resolver store.OAuthClientResolver, value string, opts ...store.QueryOption) (*models.OAuthClient, error) {
	query := []bson.M{}

	switch resolver {
	case store.OAuthClientIDResolver:
		query = append(query, bson.M{"$match": bson.M{"_id": value}})
	case store.OAuthClientClientIDResolver:
		query = append(query, bson.M{"$match": bson.M{"client_id": value}})
	}

	for _, opt := range opts {
		if err := opt(context.WithValue(ctx, "query", &query)); err != nil {
			return nil, err
		}
	}

	cursor, err := s.db.Collection("oauth_clients").Aggregate(ctx, query)
	if err != nil {
		return nil, FromMongoError(err)
	}
	defer cursor.Close(ctx)

	if !cursor.Next(ctx) {
		return nil, store.ErrNoDocuments
	}

	client := new(models.OAuthClient)
	if err := cursor.Decode(client); err != nil {
		return nil, FromMongoError(err)
	}

	return client, nil
}

func (s *Store) OAuthClientList(ctx context.Context, opts ...store.QueryOption) ([]models.OAuthClient, int, error) {
	query := []bson.M{}
	for _, opt := range opts {
		if err := opt(context.WithValue(ctx, "query", &query)); err != nil {
			return nil, 0, err
		}
	}

	count, err := CountAllMatchingDocuments(ctx, s.db.Collection("oauth_clients"), query)
	if err != nil {
		return nil, 0, FromMongoError(err)
	}

	if count == 0 {
		return []models.OAuthClient{}, 0, nil
	}

	cursor, err := s.db.Collection("oauth_clients").Aggregate(ctx, query)
	if err != nil {
		return nil, 0, FromMongoError(err)
	}
	defer cursor.Close(ctx)

	clients := make([]models.OAuthClient, 0)
	for cursor.Next(ctx) {
		client := new(models.OAuthClient)
		if err := cursor.Decode(client); err != nil {
			return nil, 0, FromMongoError(err)
		}

		clients = append(clients, *client)
	}

	return clients, count, nil
}

func (s *Store) OAuthClientDelete(ctx context.Context, client *models.OAuthClient) error {
	res, err := s.db.Collection("oauth_clients").DeleteOne(ctx, bson.M{"_id": client.ID})
	if err != nil {
		return FromMongoError(err)
	}

	if res.DeletedCount < 1 {
		return store.ErrNoDocuments
	}

	return nil
}
