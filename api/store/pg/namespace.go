package pg

import (
	"context"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/models"
)

func (s *Store) NamespaceCreate(ctx context.Context, namespace *models.Namespace) (*models.Namespace, error) {
	return nil, nil
}

func (s *Store) NamespaceList(ctx context.Context, paginator query.Paginator, filters query.Filters, opts ...store.NamespaceQueryOption) ([]models.Namespace, int, error) {
	return nil, 0, nil
}

func (s *Store) NamespaceGet(ctx context.Context, tenantID string, opts ...store.NamespaceQueryOption) (*models.Namespace, error) {
	n := new(models.Namespace)
	r := s.db.First(n, "id = ?", tenantID)

	return n, r.Error
}

func (s *Store) NamespaceGetByName(ctx context.Context, name string, opts ...store.NamespaceQueryOption) (*models.Namespace, error) {
	// TODO: unify get methods
	return nil, nil
}

func (s *Store) NamespaceGetPreferred(ctx context.Context, userID string, opts ...store.NamespaceQueryOption) (*models.Namespace, error) {
	// TODO: unify get methods
	return nil, nil
}

func (s *Store) NamespaceEdit(ctx context.Context, tenant string, changes *models.NamespaceChanges) error {
	// TODO: unify update methods
	return nil
}

func (s *Store) NamespaceUpdate(ctx context.Context, tenantID string, namespace *models.Namespace) error {
	// TODO: unify update methods
	return nil
}

func (s *Store) NamespaceDelete(ctx context.Context, tenantID string) error {
	return nil
}

// TODO: members must be an association N:N between users and namespaces now
func (s *Store) NamespaceAddMember(ctx context.Context, tenantID string, member *models.Member) error {
	return nil
}

func (s *Store) NamespaceUpdateMember(ctx context.Context, tenantID string, memberID string, changes *models.MemberChanges) error {
	return nil
}

func (s *Store) NamespaceRemoveMember(ctx context.Context, tenantID string, memberID string) error {
	return nil
}

func (s *Store) NamespaceSetSessionRecord(ctx context.Context, sessionRecord bool, tenantID string) error {
	// TODO: these methods are not used anymore
	return nil
}

func (s *Store) NamespaceGetSessionRecord(ctx context.Context, tenantID string) (bool, error) {
	// TODO: these methods are not used anymore
	return false, nil
}
