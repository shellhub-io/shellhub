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

// WithLocator sets the locator to the [APIService].
func (s *APIService) WithLocator(locator geoip.Locator) {
	s.locator = locator
}

var _ Service = (*APIService)(nil)

type Keys struct {
	PrivateKey *rsa.PrivateKey
	PublicKey  *rsa.PublicKey
}

type service struct {
	keys      Keys
	store     store.Store
	cache     cache.Cache
	validator *validator.Validator
	client    internalclient.Client
	locator   geoip.Locator
}

//go:generate mockery --name Service --filename services.go
type Service interface {
	BillingInterface
	TagsService
	DeviceService
	DeviceTags
	UserService
	SSHKeysService
	SSHKeysTagsService
	SessionService
	NamespaceService
	AuthService
	StatsService
	SetupService
	SystemService
	APIKeyService
}

func NewService(keys Keys, store store.Store, cache cache.Cache) *APIService {
	return &APIService{service: &service{
		keys, store, cache, validator.New(), internalclient.NewClient(), geoip.NewNullGeoLite(),
	}}
}
