package store

//go:generate mockery --name Store --filename store.go
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
	OAuthClientStore
	TransactionStore
	SystemStore

	Options() QueryOptions
}
