package services

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/envs"
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
	// UserDelete removes a user and cleans up related data based on the provided username, deleting
	// the namespaces the user owns and evicting the cached authorization of every device they held.
	// A failed eviction is logged and does not fail the call. It returns ErrUserNotFound,
	// ErrNamespaceNotFound when the user's namespaces cannot be read, ErrFailedDeleteUser when the
	// user row cannot be deleted, the API services' ErrForbidden for an owned tenant ID that is not a
	// valid scope, and the store's error when an owned namespace's devices cannot be listed, the namespaces cannot be
	// deleted or the user cannot depart an associated namespace, evicting nothing on any of them. ctx
	// bounds the listing, the deletion and the evictions, so a ctx cancelled after the deletion leaves
	// the cached authorizations to expire on their own.
	UserDelete(ctx context.Context, input *inputs.UserDelete) error
	// UserUpdate sets the password of the user named by the provided username and revokes every
	// token the user holds. It returns ErrUserNotFound, ErrUserPasswordInvalid, and
	// ErrFailedUpdateUser when the password cannot be saved or the tokens cannot be revoked, leaving
	// the password unchanged. Every store call runs under ctx, so a cancelled ctx surfaces as
	// ErrUserNotFound or ErrFailedUpdateUser, depending on the call it interrupts.
	UserUpdate(ctx context.Context, input *inputs.UserUpdate) error
	// UserList lists all users in the system
	UserList(ctx context.Context) ([]models.User, error)
	// UserResolve retrieves a user by their ID. It returns ErrUserNotFound
	// if no matching user exists.
	UserResolve(ctx context.Context, id string) (*models.User, error)
	// NamespaceCreate initializes a new namespace, making the specified user its owner.
	// The tenant defaults to a UUID if not provided.
	// Max device limit is based on the service's edition.
	NamespaceCreate(ctx context.Context, input *inputs.NamespaceCreate) (*models.Namespace, error)
	// NamespaceDelete deletes the namespace of the provided name and evicts the cached authorization
	// of every device it held. It returns ErrNamespaceNotFound, ErrNamespaceInstanceProtected for
	// the instance's namespace, and ErrFailedDeleteNamespace when its devices cannot be listed or
	// the store fails to delete it. A failed eviction is logged and does not fail the call. ctx bounds
	// the listing, the deletion and the evictions, so a ctx cancelled after the deletion leaves the
	// cached authorizations to expire on their own.
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
	store   store.Store
	cache   cache.Cache
	edition envs.Edition
	clock   clock.Clock
}

// Option configures the service at construction.
type Option func(*service)

// WithEdition sets the edition the namespace device limit and the connection announcement follow.
// Without it the service runs as Community.
func WithEdition(edition envs.Edition) Option {
	return func(s *service) {
		s.edition = edition
	}
}

// WithClock makes the service read the time from c. Without it the service reads the wall clock.
func WithClock(c clock.Clock) Option {
	return func(s *service) {
		s.clock = c
	}
}

// WithCache sets the cache the services evict a deleted namespace's device authorizations from.
// Without it they evict from a no-op cache, so an agent of a deleted namespace keeps authorizing
// from the server's cache until its entry expires.
func WithCache(c cache.Cache) Option {
	return func(s *service) {
		s.cache = c
	}
}

// NewService builds the admin services over store. Without WithEdition the services run as
// envs.Community, without WithClock they read the wall clock, and without WithCache they evict
// from a no-op cache.
func NewService(store store.Store, options ...Option) Services {
	s := &service{store: store, cache: cache.NewNullCache(), edition: envs.Community, clock: clock.New()}
	for _, option := range options {
		option(s)
	}

	return s
}
