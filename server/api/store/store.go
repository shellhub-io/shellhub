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
	ServiceAccountStore
	PrivateKeyStore
	StatsStore
	APIKeyStore
	InstallKeyStore
	TransactionStore
	SystemStore
	MembershipInvitationStore
	UserInvitationStore

	Options() QueryOptions
}
