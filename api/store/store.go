package store

import (
	"context"
	"errors"

	"github.com/shellhub-io/shellhub/pkg/models"
)

var (
	ErrDuplicateEmail       = errors.New("email address is already in use")
	ErrRecordNotFound       = errors.New("public key not found")
	ErrDuplicateFingerprint = errors.New("this fingerprint already exists")
	ErrNamespaceNoDocuments = errors.New("mongo: no documents in result")
	ErrDeviceNoDocuments    = errors.New("mongo: no documents in result")
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
