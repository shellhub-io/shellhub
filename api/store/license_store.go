package store

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
)

type LicenseStore interface {
	LicenseLoad(ctx context.Context) (*models.License, error)
	LicenseSave(ctx context.Context, license *models.License) error
}
