package store

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
)

type LicenseStore interface {
	LoadLicense(ctx context.Context) (*models.License, error)
	SaveLicense(ctx context.Context, license *models.License) error
}
