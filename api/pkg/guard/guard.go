package guard

import (
	"context"
	"errors"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/authorizer"
)

var ErrForbidden = errors.New("forbidden")

func getUserTypeByID(ctx context.Context, s store.Store, tenantID, id string) (string, bool) {
	user, _, err := s.UserGetByID(ctx, id, false)
	if err != nil || err == store.ErrNoDocuments {
		return "", false
	}

	namespaceUserActive, err := s.NamespaceGet(ctx, tenantID)
	if err != nil || err == store.ErrNoDocuments {
		return "", false
	}

	var userType string
	var userFound bool
	for _, member := range namespaceUserActive.Members {
		if member.ID == user.ID {
			userFound = true
			userType = member.Type

			break
		}
	}

	return userType, userFound
}

// EvaluateSubject checks if the user's type, active one, may act over another, passive one.
func EvaluateSubject(ctx context.Context, s store.Store, tenantID, activeID, typePassive string) bool {
	typeActive, ok := getUserTypeByID(ctx, s, tenantID, activeID)
	if !ok {
		return false
	}

	if typeActive == typePassive {
		return false
	}

	return authorizer.EvaluateType(typeActive, typePassive)
}

// EvaluatePermission checks if a namespace's member has the type that allows an action.
func EvaluatePermission(userType string, action int, service func() error) error {
	if !authorizer.EvaluatePermission(userType, action) {
		return ErrForbidden
	}

	return service()
}
