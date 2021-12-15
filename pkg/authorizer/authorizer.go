package authorizer

const (
	MemberRoleToken         = "token"
	MemberRoleObserver      = "observer"
	MemberRoleOperator      = "operator"
	MemberRoleAdministrator = "administrator"
	MemberRoleOwner         = "owner"
)

// GetAllMemberRoles return a list with all member roles.
// What out, when you add a new role, you need to add it to this return list.
func GetAllMemberRoles() []string {
	// The position at this slice is important to define the role priority.
	return []string{
		MemberRoleToken,
		MemberRoleObserver,
		MemberRoleOperator,
		MemberRoleAdministrator,
		MemberRoleOwner,
	}
}

func checkPermission(action int, permissions permissions) bool {
	for _, permission := range permissions {
		if permission == action {
			return true
		}
	}

	return false
}

// GetRoleCode converts a member's role to an int.
func GetRoleCode(role string) int {
	roles := GetAllMemberRoles()
	for code, roleSearch := range roles {
		if role == roleSearch {
			return code
		}
	}

	// return -1 when member role is not valid.
	return -1
}

// EvaluateRole checks if the first role has a great value than second.
func EvaluateRole(firstRole, secondRole string) bool {
	firstRoleCode := GetRoleCode(firstRole)
	secondRoleCode := GetRoleCode(secondRole)

	if firstRoleCode == -1 || secondRoleCode == -1 {
		return false
	}

	return firstRoleCode > secondRoleCode
}

// EvaluatePermission checks if the user's role has the permission to execute an action.
func EvaluatePermission(userRole string, action int) bool {
	switch userRole {
	case MemberRoleToken:
		return checkPermission(action, tokenPermission)
	case MemberRoleObserver:
		return checkPermission(action, observerPermissions)
	case MemberRoleOperator:
		return checkPermission(action, operatorPermissions)
	case MemberRoleAdministrator:
		return checkPermission(action, adminPermissions)
	case MemberRoleOwner:
		return checkPermission(action, ownerPermissions)
	}

	return false
}
