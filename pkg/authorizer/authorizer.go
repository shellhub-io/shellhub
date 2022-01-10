// Package authorizer is a hard coded package that contains the roles, actions, permissions and function related to ShellHub's API.
package authorizer

const (
	MemberRoleObserver      = "observer"
	MemberRoleOperator      = "operator"
	MemberRoleAdministrator = "administrator"
	MemberRoleOwner         = "owner"
)

// rolePermission is mapping role to its Permissions.
var rolePermission = map[string]Permissions{
	MemberRoleObserver:      observerPermissions,
	MemberRoleOperator:      operatorPermissions,
	MemberRoleAdministrator: adminPermissions,
	MemberRoleOwner:         ownerPermissions,
}

// roleCode is mapping role to its code.
// 1 has the lowest priority.
var roleCode = map[string]int{
	MemberRoleObserver:      1,
	MemberRoleOperator:      2,
	MemberRoleAdministrator: 3,
	MemberRoleOwner:         4,
}

func checkPermission(action int, permissions Permissions) bool {
	for _, permission := range permissions {
		if permission == action {
			return true
		}
	}

	return false
}

// GetRoleCode converts a member's role to an int.
func GetRoleCode(role string) int {
	code := roleCode[role]
	if code != 0 {
		return code
	}

	// return -1 when member role is not valid.
	return -1
}

// CheckRole checks if the first role has a great value than second.
func CheckRole(firstRole, secondRole string) bool {
	firstRoleCode := GetRoleCode(firstRole)
	secondRoleCode := GetRoleCode(secondRole)

	if firstRoleCode == -1 || secondRoleCode == -1 {
		return false
	}

	if firstRoleCode == secondRoleCode {
		return false
	}

	return firstRoleCode > secondRoleCode
}

// CheckPermission checks if the user's role has the permission to execute an action.
func CheckPermission(userRole string, action int) bool {
	rolePermission := rolePermission[userRole]
	if rolePermission != nil {
		return checkPermission(action, rolePermission)
	}

	return false
}
