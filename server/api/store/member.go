package store

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/models"
)

type MemberStore interface {
	NamespaceCreateMembership(ctx context.Context, sc scope.Scope, member *models.Member) error
	NamespaceUpdateMembership(ctx context.Context, sc scope.Scope, member *models.Member) error
	NamespaceDeleteMembership(ctx context.Context, sc scope.Scope, member *models.Member) error
}
