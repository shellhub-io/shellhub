package pg

import (
	"context"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/models"
)

func (pg *pg) NamespaceCreate(ctx context.Context, namespace *models.Namespace) (*models.Namespace, error) {
	return nil, nil
}

func (pg *pg) NamespaceList(ctx context.Context, opts ...store.QueryOption) ([]models.Namespace, int, error) {
	return nil, 0, nil
}

func (pg *pg) NamespaceGet(ctx context.Context, tenantID string, opts ...store.QueryOption) (*models.Namespace, error) {
	return nil, nil
}

func (pg *pg) NamespaceGetByName(ctx context.Context, name string, opts ...store.QueryOption) (*models.Namespace, error) {
	// TODO: unify get methods
	return nil, nil
}

func (pg *pg) NamespaceGetPreferred(ctx context.Context, userID string, opts ...store.QueryOption) (*models.Namespace, error) {
	// TODO: unify get methods
	return nil, nil
}

func (pg *pg) NamespaceEdit(ctx context.Context, tenant string, changes *models.NamespaceChanges) error {
	// TODO: unify update methods
	return nil
}

func (pg *pg) NamespaceUpdate(ctx context.Context, tenantID string, namespace *models.Namespace) error {
	// TODO: unify update methods
	return nil
}

func (pg *pg) NamespaceDelete(ctx context.Context, tenantID string) error {
	return nil
}

// TODO: members must be an association N:N between users and namespaces now
func (pg *pg) NamespaceAddMember(ctx context.Context, tenantID string, member *models.Member) error {
	return nil
}

func (pg *pg) NamespaceUpdateMember(ctx context.Context, tenantID string, memberID string, changes *models.MemberChanges) error {
	return nil
}

func (pg *pg) NamespaceRemoveMember(ctx context.Context, tenantID string, memberID string) error {
	return nil
}

func (pg *pg) NamespaceSetSessionRecord(ctx context.Context, sessionRecord bool, tenantID string) error {
	// TODO: these methods are not used anymore
	return nil
}

func (pg *pg) NamespaceGetSessionRecord(ctx context.Context, tenantID string) (bool, error) {
	// TODO: these methods are not used anymore
	return false, nil
}
