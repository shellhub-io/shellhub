package mongo

import (
	"context"
	"crypto/md5" //#nosec
	"encoding/hex"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

func (s *Store) TokenListAPIToken(ctx context.Context, tenantID string) ([]models.Token, error) {
	ns := new(models.Namespace)

	if err := s.db.Collection("namespaces").FindOne(ctx, bson.M{"tenant_id": tenantID}).Decode(&ns); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, store.ErrNoDocuments
		}

		return nil, err
	}

	return ns.APITokens, nil
}

func (s *Store) TokenCreateAPIToken(ctx context.Context, tenantID string) (*models.Token, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(tenantID), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	/* #nosec */
	hasher := md5.New()
	if _, err := hasher.Write(hash); err != nil {
		return nil, err
	}

	token := &models.Token{
		ID:       hex.EncodeToString(hasher.Sum(nil)),
		TenantID: tenantID,
		ReadOnly: true,
	}

	_, err = s.db.Collection("namespaces").UpdateOne(ctx, bson.M{"tenant_id": tenantID}, bson.M{"$push": bson.M{"api_tokens": token}})
	if err != nil {
		return nil, err
	}

	return token, nil
}

func (s *Store) TokenGetAPIToken(ctx context.Context, tenantID string, id string) (*models.Token, error) {
	tokens, err := s.TokenListAPIToken(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	for _, token := range tokens {
		if token.ID == id {
			return &token, nil
		}
	}

	return nil, store.ErrNoDocuments
}

func (s *Store) TokenDeleteAPIToken(ctx context.Context, tenantID string, id string) error {
	_, err := s.db.Collection("namespaces").UpdateOne(ctx, bson.M{"tenant_id": tenantID}, bson.M{"$pull": bson.M{"api_tokens": bson.M{"id": bson.M{"$eq": id}}}})
	if err != nil {
		return err
	}

	if err == mongo.ErrNoDocuments {
		return store.ErrNoDocuments
	}

	return nil
}

func (s *Store) TokenUpdateAPIToken(ctx context.Context, tenantID string, id string, request *models.APITokenUpdate) error {
	_, err := s.db.Collection("namespaces").UpdateOne(ctx, bson.M{"tenant_id": tenantID, "api_tokens.id": id}, bson.M{"$set": bson.M{"api_tokens.$.read_only": request.TokenFields.ReadOnly}})
	if err != nil {
		return err
	}

	if err == mongo.ErrNoDocuments {
		return store.ErrNoDocuments
	}

	return nil
}
