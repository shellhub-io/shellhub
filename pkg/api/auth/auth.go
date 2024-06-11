package auth

import "slices"

// Role defines a user access level.
type Role string

const (
	// RoleInvalid represents an invalid role. Any operation with this role will
	// be rejected.
	RoleInvalid Role = "N/A"
	// RoleObserver represents a namespace observer. An observer can only connect to a
	// device and retrieve device and session details.
	RoleObserver Role = "observer"
	// RoleOperator represents a namespace operator. An operator has only device-related
	// permissions, excluding the [DeviceRemove] permission. An operator also has the
	// [SessionDetails] permission.
	RoleOperator Role = "operator"
	// RoleAdministrator represents a namespace administrator. An administrator has
	// similar permissions to [RoleOwner] but cannot delete the namespace. They also do
	// not have permission for any billing-related actions.
	RoleAdministrator Role = "administrator"
	// RoleOwner represents a namespace owner. The owner has all permissions.
	RoleOwner Role = "owner"
)

// RoleFromString returns the Role corresponding to the given string.
// If the string is not a valid role, it returns [RoleInvalid].
func RoleFromString(str string) Role {
	switch str {
	case "owner":
		return RoleOwner
	case "administrator":
		return RoleAdministrator
	case "operator":
		return RoleOperator
	case "observer":
		return RoleObserver
	default:
		return RoleInvalid
	}
}

// String converts the given role to its corresponding string.
// If the string is not a valid role, it returns "N/A".
func (r Role) String() string {
	switch r {
	case RoleOwner:
		return "owner"
	case RoleAdministrator:
		return "administrator"
	case RoleOperator:
		return "operator"
	case RoleObserver:
		return "observer"
	default:
		return "N/A"
	}
}

// code converts the given role to its corresponding integer.
// If the role is not a valid one, it returns 0.
func (r Role) code() int {
	switch r {
	case RoleOwner:
		return 4
	case RoleAdministrator:
		return 3
	case RoleOperator:
		return 2
	case RoleObserver:
		return 1
	default:
		return 0
	}
}

// Permissions returns all permissions associated with the role r.
// If the role is [RoleInvalid], it returns an empty slice.
func (r Role) Permissions() []Permission {
	permissions := make([]Permission, 0)
	switch r {
	case RoleOwner:
		permissions = ownerPermissions
	case RoleAdministrator:
		permissions = adminPermissions
	case RoleOperator:
		permissions = operatorPermissions
	case RoleObserver:
		permissions = observerPermissions
	}

	return permissions
}

// HasPermission reports whether the role r has the specified permission.
func (r Role) HasPermission(permission Permission) bool {
	return slices.Contains(r.Permissions(), permission)
}

// HasAuthority reports whether the role r has greater or equal authority compared to the passive role.
// It always returns false if either role is invalid or if the passive role is [RoleOwner].
func (r Role) HasAuthority(passive Role) bool {
	return passive != RoleOwner && r.code() >= passive.code()
}
