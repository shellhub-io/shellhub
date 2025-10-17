package pg

import (
	"context"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/models"
)

func (pg *Pg) PublicKeyList(ctx context.Context, opts ...store.QueryOption) ([]models.PublicKey, int, error)

func (pg *Pg) PublicKeyGet(ctx context.Context, fingerprint string, tenantID string) (*models.PublicKey, error)

func (pg *Pg) PublicKeyCreate(ctx context.Context, key *models.PublicKey) error

func (pg *Pg) PublicKeyUpdate(ctx context.Context, publicKey *models.PublicKey) error

func (pg *Pg) PublicKeyDelete(ctx context.Context, publicKey *models.PublicKey) error
