package authorizer

const (
	MemberTypeObserver      = "observer"
	MemberTypeOperator      = "operator"
	MemberTypeAdministrator = "administrator"
	MemberTypeOwner         = "owner"
)

// GetAllMemberTypes return a list with all member types.
// What out, when you add a new type, you need to add it to this return list.
func GetAllMemberTypes() []string {
	return []string{MemberTypeObserver, MemberTypeOperator, MemberTypeAdministrator, MemberTypeOwner}
}

const (
	DeviceAccept = iota
	DeviceReject
	DeviceRemove
	DeviceConnect
	DeviceRename
	DeviceDetails

	SessionPlay
	SessionClose
	SessionRemove
	SessionDetails

	FirewallCreate
	FirewallEdit
	FirewallRemove

	PublicKeyCreate
	PublicKeyEdit
	PublicKeyRemove

	NamespaceCreate
	NamespaceRename
	NamespaceAddMember
	NamespaceRemoveMember
	NamespaceEditMember
	NamespaceEnableSessionRecord
	NamespaceDelete
)

func checkPermission(action int, permissions permissions) bool {
	for _, permission := range permissions {
		if permission == action {
			return true
		}
	}

	return false
}

// GetTypeCode converts a member's type to an int.
func GetTypeCode(memberType string) int {
	memberTypes := GetAllMemberTypes()
	for code, memberTypeSearch := range memberTypes {
		if memberType == memberTypeSearch {
			return code
		}
	}

	// return -1 when member type is not valid.
	return -1
}

// EvaluateType checks if the first type has a great value than the second.
func EvaluateType(firstType, secondType string) bool {
	firstTypeCode := GetTypeCode(firstType)
	secondTypeCode := GetTypeCode(secondType)

	if firstTypeCode == -1 || secondTypeCode == -1 {
		return false
	}

	return firstTypeCode > secondTypeCode
}

// EvaluatePermission checks if the user's type has the permission to execute an action.
func EvaluatePermission(action int, userType string) bool {
	switch userType {
	case MemberTypeObserver:
		return checkPermission(action, observerPermissions)
	case MemberTypeOperator:
		return checkPermission(action, operatorPermissions)
	case MemberTypeAdministrator:
		return checkPermission(action, adminPermissions)
	case MemberTypeOwner:
		return checkPermission(action, ownerPermissions)
	}

	return false
}
