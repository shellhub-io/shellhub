package authorizer

type permissions []int

var observerPermissions = permissions{
	DeviceConnect,
	DeviceDetails,
	SessionDetails,
	NamespaceCreate,
}

var operatorPermissions = permissions{
	DeviceAccept,
	DeviceReject,
	DeviceConnect,
	DeviceRename,
	DeviceDetails,

	SessionDetails,

	NamespaceCreate,
}

var adminPermissions = permissions{
	DeviceAccept,
	DeviceReject,
	DeviceRemove,
	DeviceConnect,
	DeviceRename,
	DeviceDetails,

	SessionPlay,
	SessionClose,
	SessionRemove,
	SessionDetails,

	FirewallCreate,
	FirewallEdit,
	FirewallRemove,

	PublicKeyCreate,
	PublicKeyEdit,
	PublicKeyRemove,

	NamespaceCreate,
	NamespaceRename,
	NamespaceAddMember,
	NamespaceRemoveMember,
	NamespaceEditMember,
	NamespaceEnableSessionRecord,
	NamespaceDelete,
}

var ownerPermissions = permissions{
	DeviceAccept,
	DeviceReject,
	DeviceRemove,
	DeviceConnect,
	DeviceRename,
	DeviceDetails,

	SessionPlay,
	SessionClose,
	SessionRemove,
	SessionDetails,

	FirewallCreate,
	FirewallEdit,
	FirewallRemove,

	PublicKeyCreate,
	PublicKeyEdit,
	PublicKeyRemove,

	NamespaceCreate,
	NamespaceRename,
	NamespaceAddMember,
	NamespaceRemoveMember,
	NamespaceEditMember,
	NamespaceEnableSessionRecord,
	NamespaceDelete,
}
