package authorizer

//go:generate go test -run TestConsolePermissionsFile -update .

// Permission is one action a role may or may not perform. The values are iota-assigned and carry
// no meaning outside the process: they are never persisted or sent on the wire, so the list can be
// reordered. The names do carry meaning: go generate writes them, with each role's set, into the
// console's permission table, so renaming one renames what the console checks.
type Permission int

// The actions a role can be granted. Grouped by resource, and deliberately finer-grained than the
// routes: one route may require several, and a role is the set it holds.
const (
	DeviceStatusUpdate Permission = iota
	DeviceUpdate
	DeviceRemove
	DeviceConnect
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
	BillingGetSubscription
	BillingGetCustomer

	APIKeyCreate
	APIKeyUpdate
	APIKeyDelete

	ProvisioningKeyCreate
	ProvisioningKeyUpdate
	ProvisioningKeyReveal
	ProvisioningKeyList

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

var observerPermissions = []Permission{
	DeviceDetails,

	SessionDetails,
}

var operatorPermissions = []Permission{
	DeviceStatusUpdate,
	DeviceConnect,
	DeviceDetails,
	DeviceUpdate,
	DeviceCustomFieldUpdate,

	TagCreate,
	TagUpdate,
	TagDelete,

	SessionDetails,
	SessionApprove,

	ProvisioningKeyList,

	SSHIdentityAdd,
}

var adminPermissions = []Permission{
	DeviceStatusUpdate,
	DeviceRemove,
	DeviceConnect,
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

	ProvisioningKeyCreate,
	ProvisioningKeyUpdate,
	ProvisioningKeyReveal,
	ProvisioningKeyList,

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
	DeviceStatusUpdate,
	DeviceRemove,
	DeviceConnect,
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
	BillingGetCustomer,

	APIKeyCreate,
	APIKeyUpdate,
	APIKeyDelete,

	ProvisioningKeyCreate,
	ProvisioningKeyUpdate,
	ProvisioningKeyReveal,
	ProvisioningKeyList,

	ConnectorDelete,
	ConnectorUpdate,
	ConnectorSet,

	TunnelsCreate,
	TunnelsDelete,

	AccessPolicyManage,

	SSHIdentityAdd,
	SSHIdentityManage,
}
