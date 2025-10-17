package pg

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
)

func (pg *Pg) PrivateKeyCreate(ctx context.Context, key *models.PrivateKey) error

func (pg *Pg) PrivateKeyGet(ctx context.Context, fingerprint string) (*models.PrivateKey, error)
