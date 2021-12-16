package authorizer

type Permissions []int

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

	NamespaceRename
	NamespaceAddMember
	NamespaceRemoveMember
	NamespaceEditMember
	NamespaceEnableSessionRecord
	NamespaceDelete

	BillingChooseDevices
	BillingAddPaymentMethod
	BillingUpdatePaymentMethod
	BillingRemovePaymentMethod
	BillingCancelSubscription
	BillingCreateSubscription
	BillingGetPaymentMethod
	BillingGetSubscription
)

var observerPermissions = Permissions{
	DeviceConnect,
	DeviceDetails,
	SessionDetails,
}

var operatorPermissions = Permissions{
	DeviceAccept,
	DeviceReject,
	DeviceConnect,
	DeviceRename,
	DeviceDetails,

	SessionDetails,
}

var adminPermissions = Permissions{
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

	NamespaceRename,
	NamespaceAddMember,
	NamespaceRemoveMember,
	NamespaceEditMember,
	NamespaceEnableSessionRecord,
}

var ownerPermissions = Permissions{
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

	NamespaceRename,
	NamespaceAddMember,
	NamespaceRemoveMember,
	NamespaceEditMember,
	NamespaceEnableSessionRecord,
	NamespaceDelete,

	BillingChooseDevices,
	BillingAddPaymentMethod,
	BillingUpdatePaymentMethod,
	BillingRemovePaymentMethod,
	BillingCancelSubscription,
	BillingCreateSubscription,
	BillingGetSubscription,
}
