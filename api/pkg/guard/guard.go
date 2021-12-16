package guard

import (
	"context"
	"errors"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/authorizer"
)

var ErrForbidden = errors.New("forbidden")

func getUserRoleByID(ctx context.Context, s store.Store, tenantID, id string) (string, bool) {
	user, _, err := s.UserGetByID(ctx, id, false)
	if err != nil || err == store.ErrNoDocuments {
		return "", false
	}

	namespaceUserActive, err := s.NamespaceGet(ctx, tenantID)
	if err != nil || err == store.ErrNoDocuments {
		return "", false
	}

	var role string
	var userFound bool
	for _, member := range namespaceUserActive.Members {
		if member.ID == user.ID {
			userFound = true
			role = member.Role

			break
		}
	}

	return role, userFound
}

// EvaluateSubject checks if the user's role, active one, may act over another, passive one.
func EvaluateSubject(ctx context.Context, s store.Store, tenantID, activeID, rolePassive string) bool {
	roleActive, ok := getUserRoleByID(ctx, s, tenantID, activeID)
	if !ok {
		return false
	}

	if roleActive == rolePassive {
		return false
	}

	return authorizer.CheckRole(roleActive, rolePassive)
}

// EvaluatePermission checks if a namespace's member has the role that allows an action.
func EvaluatePermission(role string, action int, service func() error) error {
	if !authorizer.CheckPermission(role, action) {
		return ErrForbidden
	}

	return service()
}
