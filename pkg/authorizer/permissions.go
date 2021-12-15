package authorizer

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

	BillingChooseDevices
	BillingAddPaymentMethod
	BillingUpdatePaymentMethod
	BillingRemovePaymentMethod
	BillingCancelSubscription
	BillingCreateSubscription
	BillingGetPaymentMethod
	BillingGetSubscription

	TokenList
	TokenCreate
	TokenGet
	TokenEdit
	TokenRemove
)

type permissions []int

var tokenPermission = permissions{
	DeviceDetails,
	SessionDetails,
}

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

	TokenList,
	TokenCreate,
	TokenGet,
	TokenEdit,
	TokenRemove,
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

	BillingChooseDevices,
	BillingAddPaymentMethod,
	BillingUpdatePaymentMethod,
	BillingRemovePaymentMethod,
	BillingCancelSubscription,
	BillingCreateSubscription,
	BillingGetSubscription,

	TokenList,
	TokenCreate,
	TokenGet,
	TokenEdit,
	TokenRemove,
}
