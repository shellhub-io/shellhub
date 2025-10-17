package pg

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
)

func (pg *Pg) GetStats(ctx context.Context) (*models.Stats, error)
