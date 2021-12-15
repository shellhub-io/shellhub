package authorizer

const (
	MemberRoleToken         = "token"
	MemberRoleObserver      = "observer"
	MemberRoleOperator      = "operator"
	MemberRoleAdministrator = "administrator"
	MemberRoleOwner         = "owner"
)

// RolePermission is mapping role to its permissions.
var RolePermission = map[string]permissions{
	MemberRoleToken:         tokenPermission,
	MemberRoleObserver:      observerPermissions,
	MemberRoleOperator:      operatorPermissions,
	MemberRoleAdministrator: adminPermissions,
	MemberRoleOwner:         ownerPermissions,
}

// RoleCode is mapping role to is code.
// 1 has the lowest priority.
var RoleCode = map[string]int{
	MemberRoleToken:         1,
	MemberRoleObserver:      2,
	MemberRoleOperator:      3,
	MemberRoleAdministrator: 4,
	MemberRoleOwner:         5,
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
	code := RoleCode[role]
	if code != 0 {
		return code
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
	rolePermission := RolePermission[userRole]
	if rolePermission != nil {
		return checkPermission(action, rolePermission)
	}

	return false
}
