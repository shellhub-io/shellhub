package store

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
)

type MemberStore interface {
	NamespaceCreateMembership(ctx context.Context, tenantID string, member *models.Member) error
	NamespaceUpdateMembership(ctx context.Context, tenantID string, member *models.Member) error
	NamespaceDeleteMembership(ctx context.Context, tenantID string, member *models.Member) error
}
