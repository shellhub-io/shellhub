package store

// Store is everything the server persists, composed from the per-resource stores. One backing
// implementation supplies all of them, so an operation may span resources in a transaction.
type Store interface {
	TagsStore
	DeviceStore
	SessionStore
	UserStore
	NamespaceStore
	MemberStore
	PublicKeyStore
	AccessPolicyStore
	SSHIdentityStore
	SSHApprovalStore
	PrivateKeyStore
	StatsStore
	APIKeyStore
	InstanceAPIKeyStore
	ProvisioningKeyStore
	TransactionStore
	SystemStore
	MembershipInvitationStore
	UserInvitationStore

	Options() QueryOptions
}
