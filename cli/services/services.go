package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/models"
)

const (
	MaxNumberNamespacesCommunity = -1 // MaxNumberNamespacesCommunity defines a unlimited number of namespaces for communitty environment.
	MaxNumberDevicesLimited      = 3  // MaxNumberDevicesLimited defines the limit of devices for cloud environment.
	MaxNumberDevicesUnlimited    = -1 // MaxNumberDevicesUnlimited defines a unlimited number of devices for enterprise and community environment.
)

type Services interface {
	// UserCreate adds a new user based on the provided user's data. This method validates data and
	// checks for conflicts.
	UserCreate(ctx context.Context, username, password, email string) (*models.User, error)
	// UserDelete removes a user and cleans up related data based on the provided username.
	UserDelete(ctx context.Context, username string) error
	// UserUpdate updates a user's data based on the provided username.
	UserUpdate(ctx context.Context, username, password string) error
	// NamespaceCreate initializes a new namespace, making the specified user its owner.
	// The tenant defaults to a UUID if not provided.
	// Max device limit is based on the envs.IsCloud() setting.
	NamespaceCreate(ctx context.Context, namespace, username, tenant string) (*models.Namespace, error)
	// NamespaceAddMember adds a new member with a specified role to a namespace.
	NamespaceAddMember(ctx context.Context, username, namespace, role string) (*models.Namespace, error)
	// NamespaceRemoveMember removes a member from a namespace.
	NamespaceRemoveMember(ctx context.Context, username, namespace string) (*models.Namespace, error)
	// NamespaceDelete deletes a namespace based on the provided namespace name.
	NamespaceDelete(ctx context.Context, namespace string) error
}

// service is an internal struct that implements the Services interface.
// It contains a store, which provides a mechanism to interact with the data store.
type service struct {
	store store.Store
}

// NewService creates and returns a new instance of the service with the provided store.
func NewService(store store.Store) Services {
	return &service{store}
}

// hashPassword computes the sha256 hash of the given password and returns
// the hex encoded string representation of the hash.
func hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))

	return hex.EncodeToString(hash[:])
}

// normalizeField converts the provided string data to lowercase.
func normalizeField(data string) string {
	return strings.ToLower(data)
}
