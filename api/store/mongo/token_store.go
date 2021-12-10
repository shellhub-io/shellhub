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

func (s *Store) TokenList(ctx context.Context, tenantID string) ([]models.Token, error) {
	ns := new(models.Namespace)

	if err := s.db.Collection("namespaces").FindOne(ctx, bson.M{"tenant_id": tenantID}).Decode(&ns); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, store.ErrNoDocuments
		}

		return nil, err
	}

	return ns.Tokens, nil
}

func (s *Store) TokenCreate(ctx context.Context, tenantID string) (*models.Token, error) {
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

	_, err = s.db.Collection("namespaces").UpdateOne(ctx, bson.M{"tenant_id": tenantID}, bson.M{"$push": bson.M{"tokens": token}})
	if err != nil {
		return nil, err
	}

	return token, nil
}

func (s *Store) TokenGet(ctx context.Context, tenantID string, id string) (*models.Token, error) {
	tokens, err := s.TokenList(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	var token *models.Token
	for _, t := range tokens {
		if t.ID == id {
			*token = t
			break
		}
	}

	return token, nil
}

func (s *Store) TokenDelete(ctx context.Context, tenantID string, id string) error {
	_, err := s.db.Collection("namespaces").UpdateOne(ctx, bson.M{"tenant_id": tenantID}, bson.M{"$pull": bson.M{"tokens": bson.M{"id": bson.M{"$eq": id}}}})
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return err
		}

		return err
	}

	return nil
}

func (s *Store) TokenUpdate(ctx context.Context, tenantID string, id string, readOnly bool) error {
	_, err := s.db.Collection("namespaces").UpdateOne(ctx, bson.M{"tenant_id": tenantID, "tokens.id": id}, bson.M{"$set": bson.M{"tokens.$.read_only": readOnly}})
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return err
		}

		return err
	}

	return nil
}
