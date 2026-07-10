package store

type Store interface {
	TagsStore
	DeviceStore
	SessionStore
	UserStore
	NamespaceStore
	MemberStore
	PublicKeyStore
	PrivateKeyStore
	StatsStore
	APIKeyStore
	TransactionStore
	SystemStore
	MembershipInvitationStore
	UserInvitationStore

	Options() QueryOptions
}
