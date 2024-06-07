// Package guard is a helper to work around permissions on ShellHub API.
package guard

import (
	"github.com/shellhub-io/shellhub/pkg/models"
)

const (
	RoleObserver      = "observer"
	RoleOperator      = "operator"
	RoleAdministrator = "administrator"
	RoleOwner         = "owner"
)

// RoleInvalidCode is a role code for invalid role.
const RoleInvalidCode = -1

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

// GetRoleCode converts a models.Member's role string to a role code. If the role is not found in Roles, it returns RoleInvalidCode.
func GetRoleCode(role string) int {
	code, ok := Roles[role]
	if !ok {
		// return RoleInvalidCode when member's role is not valid.
		return RoleInvalidCode
	}

	return code
}

// HasAuthority reports whether the active role has greater or equal authority compared to the passive role.
// It returns false if either role is invalid or if the passive role is [RoleOwner].
func HasAuthority(active, passive string) bool {
	if passive == RoleOwner {
		return false
	}

	activeCode := GetRoleCode(active)
	passiveCode := GetRoleCode(passive)

	if activeCode == RoleInvalidCode || passiveCode == RoleInvalidCode {
		return false
	}

	return activeCode >= passiveCode
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
	member, ok := namespace.FindMember(userID)
	if !ok {
		return ErrForbidden
	}

	return EvaluatePermission(member.Role, action, callback)
}
