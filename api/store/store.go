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

	// Save saves a model or a list of models. It returns an error if any.
	Save(ctx context.Context, models ...any) error

	// Delete deletes a model or a list of models. It returns an error if any.
	Delete(ctx context.Context, models ...any) error
}
