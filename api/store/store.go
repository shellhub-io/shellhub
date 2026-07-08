package store

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
