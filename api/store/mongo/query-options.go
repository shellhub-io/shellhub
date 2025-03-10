package mongo

import (
	"context"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/models"
)

func (s *Store) Options() store.QueryOptions {
	return s.options
}

func (*queryOptions) CountAcceptedDevices() store.NamespaceQueryOption {
	return func(ctx context.Context, ns *models.Namespace) error {
		return nil
	}
}

func (*queryOptions) EnrichMembersData() store.NamespaceQueryOption {
	return func(ctx context.Context, ns *models.Namespace) error {
		return nil
	}
}
