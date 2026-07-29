package authorizer

type Permission int

const (
	DeviceAccept Permission = iota
	DeviceReject
	DeviceUpdate
	DeviceRemove
	DeviceConnect
	DeviceRename
	DeviceDetails
	DeviceCustomFieldUpdate

	TagCreate
	TagUpdate
	TagDelete

	SessionPlay
	SessionClose
	SessionRemove
	SessionDetails
	SessionApprove

	FirewallCreate
	FirewallEdit
	FirewallRemove

	PublicKeyCreate
	PublicKeyEdit
	PublicKeyRemove

	NamespaceUpdate
	NamespaceAddMember
	NamespaceRemoveMember
	NamespaceEditMember
	NamespaceEnableSessionRecord
	NamespaceDelete

	BillingCreateCustomer
	BillingChooseDevices
	BillingAddPaymentMethod
	BillingUpdatePaymentMethod
	BillingRemovePaymentMethod
	BillingCancelSubscription
	BillingCreateSubscription
	BillingGetPaymentMethod
	BillingGetSubscription

	APIKeyCreate
	APIKeyUpdate
	APIKeyDelete

	InstallKeyCreate
	InstallKeyUpdate
	InstallKeyReveal
	InstallKeyList

	ConnectorDelete
	ConnectorUpdate
	ConnectorSet

	TunnelsCreate
	TunnelsDelete

	AccessPolicyManage

	// SSHIdentityAdd allows adding and managing one's own SSH identities.
	// Owner/admin/operator.
	SSHIdentityAdd
	// SSHIdentityManage allows viewing and revoking any member's SSH identities
	// in the namespace (offboarding). Owner/admin only.
	SSHIdentityManage
)

// servicePermissions is intentionally empty: a service account has no management
// permissions (see [RoleService]).
var servicePermissions = []Permission{}

var observerPermissions = []Permission{
	DeviceConnect,
	DeviceDetails,

	SessionDetails,
}

var operatorPermissions = []Permission{
	DeviceAccept,
	DeviceReject,
	DeviceConnect,
	DeviceRename,
	DeviceDetails,
	DeviceUpdate,
	DeviceCustomFieldUpdate,

	TagCreate,
	TagUpdate,
	TagDelete,

	SessionDetails,
	SessionApprove,

	SSHIdentityAdd,
}

var adminPermissions = []Permission{
	DeviceAccept,
	DeviceReject,
	DeviceRemove,
	DeviceConnect,
	DeviceRename,
	DeviceDetails,
	DeviceUpdate,
	DeviceCustomFieldUpdate,

	TagCreate,
	TagUpdate,
	TagDelete,

	SessionPlay,
	SessionClose,
	SessionRemove,
	SessionDetails,
	SessionApprove,

	FirewallCreate,
	FirewallEdit,
	FirewallRemove,

	PublicKeyCreate,
	PublicKeyEdit,
	PublicKeyRemove,

	NamespaceUpdate,
	NamespaceAddMember,
	NamespaceRemoveMember,
	NamespaceEditMember,
	NamespaceEnableSessionRecord,

	APIKeyCreate,
	APIKeyUpdate,
	APIKeyDelete,

	InstallKeyCreate,
	InstallKeyUpdate,
	InstallKeyReveal,
	InstallKeyList,

	ConnectorDelete,
	ConnectorUpdate,
	ConnectorSet,

	TunnelsCreate,
	TunnelsDelete,

	AccessPolicyManage,

	SSHIdentityAdd,
	SSHIdentityManage,
}

var ownerPermissions = []Permission{
	DeviceAccept,
	DeviceReject,
	DeviceRemove,
	DeviceConnect,
	DeviceRename,
	DeviceDetails,
	DeviceUpdate,
	DeviceCustomFieldUpdate,

	TagCreate,
	TagUpdate,
	TagDelete,

	SessionPlay,
	SessionClose,
	SessionRemove,
	SessionDetails,
	SessionApprove,

	FirewallCreate,
	FirewallEdit,
	FirewallRemove,

	PublicKeyCreate,
	PublicKeyEdit,
	PublicKeyRemove,

	NamespaceUpdate,
	NamespaceAddMember,
	NamespaceRemoveMember,
	NamespaceEditMember,
	NamespaceEnableSessionRecord,
	NamespaceDelete,

	BillingCreateCustomer,
	BillingChooseDevices,
	BillingAddPaymentMethod,
	BillingUpdatePaymentMethod,
	BillingRemovePaymentMethod,
	BillingCancelSubscription,
	BillingCreateSubscription,
	BillingGetSubscription,

	APIKeyCreate,
	APIKeyUpdate,
	APIKeyDelete,

	InstallKeyCreate,
	InstallKeyUpdate,
	InstallKeyReveal,
	InstallKeyList,

	ConnectorDelete,
	ConnectorUpdate,
	ConnectorSet,

	TunnelsCreate,
	TunnelsDelete,

	AccessPolicyManage,

	SSHIdentityAdd,
	SSHIdentityManage,
}
