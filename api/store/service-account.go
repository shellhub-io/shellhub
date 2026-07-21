package store

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
)

type ServiceAccountStore interface {
	// ServiceAccountList returns the service accounts (service-typed members) of the
	// namespace, ordered by creation time. The returned identities slice is left empty
	// for the caller to populate.
	ServiceAccountList(ctx context.Context, tenantID string) ([]models.ServiceAccount, int, error)
}
