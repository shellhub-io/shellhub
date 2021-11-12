package guard

import (
	"context"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/authorizer"
)

func getTypeByID(ctx context.Context, s store.Store, tenantID, id string) (string, bool) {
	user, _, err := s.UserGetByID(ctx, id, false)
	if err != nil || err == store.ErrNoDocuments {
		return "", false
	}

	namespaceUserActive, err := s.NamespaceGet(ctx, tenantID)
	if err != nil || err == store.ErrNoDocuments {
		return "", false
	}

	var userType string
	for _, member := range namespaceUserActive.Members {
		if member.ID == user.ID {
			userType = member.Type
		}
	}
	if userType == "" {
		return "", false
	}

	return userType, true
}

// EvaluateSubject checks if the user's type, active one, may act over another, passive one.
func EvaluateSubject(ctx context.Context, s store.Store, tenantID, activeID, typePassive string) bool {
	typeActive, ok := getTypeByID(ctx, s, tenantID, activeID)
	if !ok {
		return false
	}

	if typeActive == typePassive {
		return false
	}

	return authorizer.EvaluateType(typeActive, typePassive)
}

// EvaluatePermission checks if a namespace's member has the type what allow an action.
func EvaluatePermission(ctx context.Context, s store.Store, tenantID, userID string, action int) bool {
	userType, ok := getTypeByID(ctx, s, tenantID, userID)
	if !ok {
		return false
	}

	return authorizer.EvaluatePermission(action, userType)
}
