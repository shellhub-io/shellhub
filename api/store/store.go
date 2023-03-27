package store

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
	SlotStore
}
