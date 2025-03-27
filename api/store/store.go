package store

import "context"

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
	SystemStore

	Options() QueryOptions

	Save(ctx context.Context, models ...any) error
	Delete(ctx context.Context, models ...any) error
}
