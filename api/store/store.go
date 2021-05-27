package store

import (
	"context"
	"errors"

	"github.com/shellhub-io/shellhub/pkg/models"
)

var (
	ErrDuplicate      = errors.New("duplicate")
	ErrDuplicateUser  = errors.New("user already exists")
	ErrDuplicateEmail = errors.New("email address is already in use")
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
