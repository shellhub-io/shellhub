package store

//go:generate mockery --name Store --filename store.go
type Store interface {
	TagsStore
	DeviceStore
	DeviceTagsStore
	SessionStore
	UserStore
	NamespaceStore
	PublicKeyStore
	PublicKeyTagsStore
	PrivateKeyStore
	StatsStore
	APIKeyStore
	TransactionStore
}
