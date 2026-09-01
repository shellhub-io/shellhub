package services

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/admin/inputs"
	"github.com/shellhub-io/shellhub/server/api/store"
)

// The per-edition ceilings on namespaces and devices. A negative value means no limit.
const (
	MaxNumberNamespacesCommunity = -1 // MaxNumberNamespacesCommunity defines a unlimited number of namespaces for communitty environment.
	MaxNumberDevicesLimited      = 3  // MaxNumberDevicesLimited defines the limit of devices for cloud environment.
	MaxNumberDevicesUnlimited    = -1 // MaxNumberDevicesUnlimited defines a unlimited number of devices for enterprise and community environment.
)

// NamespaceResolver names the field an admin command looks a namespace up by.
type NamespaceResolver string

// The fields a namespace can be resolved by: its human-facing name, or its tenant ID.
const (
	NamespaceResolverName     NamespaceResolver = "name"
	NamespaceResolverTenantID NamespaceResolver = "tenant-id"
)

// Services is the API behind the admin CLI. It is deliberately separate from the REST API's
// services: these operations run as the instance operator and are not reachable over HTTP.
type Services interface {
	// UserCreate adds a new user based on the provided user's data. This method validates data and
	// checks for conflicts.
	UserCreate(ctx context.Context, input *inputs.UserCreate) (*models.User, error)
	// UserDelete removes a user and cleans up related data based on the provided username.
	UserDelete(ctx context.Context, input *inputs.UserDelete) error
	// UserUpdate updates a user's data based on the provided username.
	UserUpdate(ctx context.Context, input *inputs.UserUpdate) error
	// UserList lists all users in the system
	UserList(ctx context.Context) ([]models.User, error)
	// UserResolve retrieves a user by their ID. It returns ErrUserNotFound
	// if no matching user exists.
	UserResolve(ctx context.Context, id string) (*models.User, error)
	// NamespaceCreate initializes a new namespace, making the specified user its owner.
	// The tenant defaults to a UUID if not provided.
	// Max device limit is based on the envs.IsCloud() setting.
	NamespaceCreate(ctx context.Context, input *inputs.NamespaceCreate) (*models.Namespace, error)
	// NamespaceDelete deletes a namespace based on the provided namespace name.
	NamespaceDelete(ctx context.Context, input *inputs.NamespaceDelete) error
	// NamespaceList retrieves all namespaces available to the user
	NamespaceList(ctx context.Context) ([]models.Namespace, error)
	// NamespaceResolve retrieves a namespace using the specified resolver
	NamespaceResolve(ctx context.Context, resolver NamespaceResolver, value string) (*models.Namespace, error)
	// NamespaceAddMember adds a new member with a specified role to a namespace.
	NamespaceAddMember(ctx context.Context, input *inputs.MemberAdd) (*models.Namespace, error)
	// NamespaceRemoveMember removes a member from a namespace.
	NamespaceRemoveMember(ctx context.Context, input *inputs.MemberRemove) (*models.Namespace, error)
	// NamespaceDeviceCounts returns the actual device counts for a namespace.
	NamespaceDeviceCounts(ctx context.Context, tenantID string) (*models.Stats, error)
}

type service struct {
	store store.Store
}

// NewService creates and returns a new instance of the service with the provided store.
func NewService(store store.Store) Services {
	return &service{store}
}
