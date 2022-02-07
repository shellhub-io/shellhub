package store

type Store interface {
	DeviceStore
	DeviceTagsStore
	SessionStore
	UserStore
	FirewallStore
	NamespaceStore
	PublicKeyStore
	PrivateKeyStore
	LicenseStore
	StatsStore
	BillingStore
}
