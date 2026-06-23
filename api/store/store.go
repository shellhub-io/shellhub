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
	ConnectionStore
	KnownHostStore
	TransactionStore
	SystemStore

	Options() QueryOptions
}
