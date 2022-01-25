// Package guard is a helper package to evaluate question about members in ShellHub.
package guard

import (
	"errors"

	"github.com/shellhub-io/shellhub/pkg/authorizer"
	"github.com/shellhub-io/shellhub/pkg/models"
)

var ErrForbidden = errors.New("forbidden")

// CheckMember checks if a user is a namespace's member.
func CheckMember(namespace *models.Namespace, id string) (*models.Member, bool) {
	var memberFound models.Member
	for _, memberSearch := range namespace.Members {
		if memberSearch.ID == id {
			memberFound = memberSearch

			break
		}
	}

	if memberFound.ID == "" || memberFound.Role == "" {
		return nil, false
	}

	return &memberFound, true
}

// CheckRole checks if a member from a namespace can act over other with a specif role.
func CheckRole(roleActive, rolePassive string) bool {
	return authorizer.CheckRole(roleActive, rolePassive)
}

// EvaluatePermission checks if a namespace's member has the role that allows an action.
func EvaluatePermission(role string, action int, callback func() error) error {
	if !authorizer.CheckPermission(role, action) {
		return ErrForbidden
	}

	return callback()
}

func EvaluateNamespace(namespace *models.Namespace, userID string, action int, callback func() error) error {
	mb, ok := CheckMember(namespace, userID)
	if !ok {
		return ErrForbidden
	}

	if !authorizer.CheckPermission(mb.Role, action) {
		return ErrForbidden
	}

	return callback()
}
