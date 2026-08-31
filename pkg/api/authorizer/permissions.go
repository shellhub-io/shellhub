package authorizer

// Permission is one action a role may or may not perform. The values are iota-assigned and carry
// no meaning outside the process: they are never persisted or sent on the wire, so the list can be
// reordered.
type Permission int

// The actions a role can be granted. Grouped by resource, and deliberately finer-grained than the
// routes: one route may require several, and a role is the set it holds.
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
