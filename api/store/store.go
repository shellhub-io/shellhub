package store

import (
	"context"
	"errors"

	"github.com/shellhub-io/shellhub/pkg/models"
)

var (
	ErrDuplicateEmail = errors.New("email address is already in use")
	ErrDuplicate      = errors.New("this fingerprint already exists")
	ErrNoDocuments    = errors.New("mongo: no documents in result")
	ErrInvalidHex     = errors.New("the provided hex string is not a valid ObjectID")
)

type Store interface {
	DeviceStore
	SessionStore
	UserStore
	FirewallStore
	NamespaceStore
	PublicKeyStore
	PrivateKeyStore
	LicenseStore
	GetStats(ctx context.Context) (*models.Stats, error)
}
