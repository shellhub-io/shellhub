package pg

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/models"
)

func (pg *pg) PublicKeyCreate(ctx context.Context, key *models.PublicKey) error {
	return nil
}

func (pg *pg) PublicKeyList(ctx context.Context, paginator query.Paginator) ([]models.PublicKey, int, error) {
	return nil, 0, nil
}

func (pg *pg) PublicKeyGet(ctx context.Context, fingerprint string, tenantID string) (*models.PublicKey, error) {
	return nil, nil
}

func (pg *pg) PublicKeyUpdate(ctx context.Context, fingerprint string, tenantID string, key *models.PublicKeyUpdate) (*models.PublicKey, error) {
	return nil, nil
}

func (pg *pg) PublicKeyDelete(ctx context.Context, fingerprint string, tenantID string) error {
	return nil
}

func (pg *pg) PublicKeyPushTag(ctx context.Context, tenant, fingerprint, tag string) error {
	// TODO: refactor tags entirely
	return nil
}

func (pg *pg) PublicKeyPullTag(ctx context.Context, tenant, fingerprint, tag string) error {
	// TODO: refactor tags entirely
	return nil
}

func (pg *pg) PublicKeySetTags(ctx context.Context, tenant, fingerprint string, tags []string) (matchedCount int64, updatedCount int64, err error) {
	// TODO: refactor tags entirely
	return 0, 0, nil
}

func (pg *pg) PublicKeyBulkRenameTag(ctx context.Context, tenant, currentTag, newTag string) (updatedCount int64, err error) {
	// TODO: refactor tags entirely
	return 0, nil
}

func (pg *pg) PublicKeyBulkDeleteTag(ctx context.Context, tenant, tag string) (updatedCount int64, err error) {
	// TODO: refactor tags entirely
	return 0, nil
}

func (pg *pg) PublicKeyGetTags(ctx context.Context, tenant string) (tag []string, size int, err error) {
	// TODO: refactor tags entirely
	return nil, 0, nil
}
