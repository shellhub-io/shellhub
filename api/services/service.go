package services

import (
	"crypto/rsa"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/internalclient"
	"github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/geoip"
	"github.com/shellhub-io/shellhub/pkg/validator"
)

type APIService struct {
	*service
}

var _ Service = (*APIService)(nil)

type service struct {
	store     store.Store
	privKey   *rsa.PrivateKey
	pubKey    *rsa.PublicKey
	cache     cache.Cache
	client    internalclient.Client
	locator   geoip.Locator
	validator *validator.Validator
	billing   BillingProvider
}

//go:generate mockery --name Service --filename services.go
type Service interface {
	TagsService
	DeviceService
	UserService
	SSHKeysService
	SessionService
	NamespaceService
	MemberService
	AuthService
	StatsService
	SetupService
	SystemService
	APIKeyService

	// Store returns the underlying store instance.
	//
	// IMPORTANT: Extensions should use Store() ONLY for:
	// 1. Creating their own service layers that need database access
	// 2. Querying data not exposed by core Service methods
	//
	// Extensions MUST NOT:
	// 1. Modify core data (devices, users, namespaces) directly via Store()
	// 2. Bypass core business logic or validation rules
	// 3. Create transactions that span both core and extension operations
	//
	// Violating these rules may cause data inconsistency or break core functionality.
	Store() store.Store
}

type Option func(service *APIService)

func WithLocator(locator geoip.Locator) Option {
	return func(service *APIService) {
		service.locator = locator
	}
}

func WithBilling(billing BillingProvider) Option {
	return func(service *APIService) {
		service.billing = billing
	}
}

func NewService(store store.Store, privKey *rsa.PrivateKey, pubKey *rsa.PublicKey, cache cache.Cache, c internalclient.Client, options ...Option) *APIService {
	if privKey == nil || pubKey == nil {
		var err error
		privKey, pubKey, err = LoadKeys()
		if err != nil {
			panic(err)
		}
	}

	service := &APIService{
		service: &service{
			store,
			privKey,
			pubKey,
			cache,
			c,
			geoip.NewNullGeoLite(),
			validator.New(),
			nil, // billing (injected via WithBilling option)
		},
	}

	for _, option := range options {
		option(service)
	}

	return service
}

// Store returns the underlying store instance.
// This allows route extensions (enterprise/cloud) to access the same database connection.
func (s *APIService) Store() store.Store {
	return s.store
}
