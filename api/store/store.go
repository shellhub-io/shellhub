package store

//go:generate mockery --name Store --filename store.go
type Store interface {
	AnnouncementsStore
	TagsStore
	DeviceStore
	DeviceTagsStore
	SessionStore
	UserStore
	FirewallStore
	FirewallTagsStore
	NamespaceStore
	PublicKeyStore
	PublicKeyTagsStore
	PrivateKeyStore
	LicenseStore
	StatsStore
}
