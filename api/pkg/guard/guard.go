// Package guard is a helper to work around permissions on ShellHub API.
package guard

import (
	"errors"

	"github.com/shellhub-io/shellhub/pkg/models"
)

const (
	RoleObserver      = "observer"
	RoleOperator      = "operator"
	RoleAdministrator = "administrator"
	RoleOwner         = "owner"
)

const (
	// RoleObserverCode is a role code for observer.
	RoleObserverCode = iota + 1
	// RoleOperatorCode is a role code for operator.
	RoleOperatorCode
	// RoleAdministratorCode is a role code for administrator.
	RoleAdministratorCode
	// RoleOwnerCode is a role code for owner.
	RoleOwnerCode
)

// Roles maps all roles to its code.
var Roles = map[string]int{
	RoleObserver:      RoleObserverCode,
	RoleOperator:      RoleOperatorCode,
	RoleAdministrator: RoleAdministratorCode,
	RoleOwner:         RoleOwnerCode,
}

// RolePermissions maps roles to its Permissions. It is used to check if a models.Member has permission to do something.
var RolePermissions = map[string]Permissions{
	RoleObserver:      observerPermissions,
	RoleOperator:      operatorPermissions,
	RoleAdministrator: adminPermissions,
	RoleOwner:         ownerPermissions,
}

var ErrForbidden = errors.New("forbidden")

// CheckMember checks if a models.User's ID is a models.Namespace's member. A models.User is a member if its ID is in
// the models.Namespace's members list.
func CheckMember(namespace *models.Namespace, id string) (*models.Member, bool) {
	var found models.Member
	for _, member := range namespace.Members {
		if member.ID == id {
			found = member

			break
		}
	}

	if found.ID == "" || found.Role == "" {
		return nil, false
	}

	return &found, true
}

// GetRoleCode converts a models.Member's role string to a role code. If the role is not found in Roles, it returns -1.
func GetRoleCode(role string) int {
	code, ok := Roles[role]
	if !ok {
		// return -1 when member role is not valid.
		return -1
	}

	return code
}

// CheckRole checks if a models.Member's role from a models.Namespace can act over the other. Active is the member's role
// from who is acting, and passive is the member who is being acted. Active and passive roles must be members of the
// same models.Namespace.
//
// If active or passive is an invalid member, it returns false. If active and passive are equal, it returns false too.
//
// Active and passive must be one of the following: RoleObserver, RoleOperator,RoleAdmin or RoleOwner.
func CheckRole(active, passive string) bool {
	first := GetRoleCode(active)
	second := GetRoleCode(passive)

	if first == -1 || second == -1 {
		return false
	}

	if first == second {
		return false
	}

	return first > second
}

// EvaluatePermission checks if a models.Namespace's member has the role that allows an action. Each role has a list of
// allowed actions.
//
// Role is the member's role from who is acting, Action is the action that is being performed and callback is a function
// to be called if the action is allowed.
func EvaluatePermission(role string, action int, callback func() error) error {
	check := func(action int, permissions Permissions) bool {
		for _, permission := range permissions {
			if permission == action {
				return true
			}
		}

		return false
	}

	permission, ok := RolePermissions[role]
	if !ok {
		return ErrForbidden
	}

	if !check(action, permission) {
		return ErrForbidden
	}

	return callback()
}

func EvaluateNamespace(namespace *models.Namespace, userID string, action int, callback func() error) error {
	member, ok := CheckMember(namespace, userID)
	if !ok {
		return ErrForbidden
	}

	return EvaluatePermission(member.Role, action, callback)
}
