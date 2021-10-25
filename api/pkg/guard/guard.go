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

func getTypeUsername(ctx context.Context, s store.Store, tenantID, username string) (string, bool) {
	user, err := s.UserGetByUsername(ctx, username)
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

func EvaluateSubjectWithUsername(ctx context.Context, s store.Store, tenantID, activeID, passiveUsername string) bool {
	typeActive, ok := getTypeByID(ctx, s, tenantID, activeID)
	if !ok {
		return false
	}
	typePassive, _ := getTypeUsername(ctx, s, tenantID, passiveUsername)
	if !ok {
		return false
	}

	userTypeCode := authorizer.GetTypeCode(typeActive)
	passiveTypeCode := authorizer.GetTypeCode(typePassive)
	if userTypeCode == -1 || passiveTypeCode == -1 {
		return false
	}

	if userTypeCode < passiveTypeCode {
		return false
	}

	return true
}

func EvaluateSubjectType(ctx context.Context, s store.Store, tenantID, activeID, typePassive string) bool {
	typeActive, ok := getTypeByID(ctx, s, tenantID, activeID)
	if !ok {
		return false
	}

	userTypeCode := authorizer.GetTypeCode(typeActive)
	passiveTypeCode := authorizer.GetTypeCode(typePassive)
	if userTypeCode == -1 || passiveTypeCode == -1 {
		return false
	}

	if userTypeCode <= passiveTypeCode {
		return false
	}

	return true
}

// EvaluatePermission checks if a namespace's member has the type what allow an action.
func EvaluatePermission(ctx context.Context, s store.Store, tenantID, userID string, action int) bool {
	user, _, err := s.UserGetByID(ctx, userID, false)
	if err != nil || err == store.ErrNoDocuments {
		return false
	}

	namespace, err := s.NamespaceGet(ctx, tenantID)
	if err != nil || err == store.ErrNoDocuments {
		return false
	}

	var userType string
	for _, member := range namespace.Members {
		if member.ID == user.ID {
			userType = member.Type
		}
	}
	if userType == "" {
		return false
	}

	return authorizer.Evaluate(action, userType)
}
