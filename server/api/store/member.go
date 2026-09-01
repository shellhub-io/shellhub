package store

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/models"
)

// MemberStore persists namespace membership: who belongs to a namespace, and in what role.
type MemberStore interface {
	NamespaceCreateMembership(ctx context.Context, sc scope.Scope, member *models.Member) error
	NamespaceUpdateMembership(ctx context.Context, sc scope.Scope, member *models.Member) error
	NamespaceDeleteMembership(ctx context.Context, sc scope.Scope, member *models.Member) error
}
