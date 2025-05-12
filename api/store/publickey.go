package store

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
)

type PublicKeyIdent string

const (
	PublicKeyIdentFingerprint PublicKeyIdent = "fingerprint"
)

type PublicKeyStore interface {
	PublicKeyCreate(ctx context.Context, publicKey *models.PublicKey) (string, error)
	PublicKeyList(ctx context.Context, opts ...QueryOption) ([]models.PublicKey, int, error)
	PublicKeyGet(ctx context.Context, ident PublicKeyIdent, val string, tenantID string) (*models.PublicKey, error)
	PublicKeySave(ctx context.Context, publicKey *models.PublicKey) error
	PublicKeyDelete(ctx context.Context, publicKey *models.PublicKey) error
}
