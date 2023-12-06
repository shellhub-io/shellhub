package services

import (
	"crypto/rsa"

	"github.com/shellhub-io/shellhub/api/store"
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
	client    interface{}
	locator   geoip.Locator
	validator *validator.Validator
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
}

func NewService(store store.Store, privKey *rsa.PrivateKey, pubKey *rsa.PublicKey, cache cache.Cache, c interface{}, l geoip.Locator) *APIService {
	if privKey == nil || pubKey == nil {
		var err error
		privKey, pubKey, err = LoadKeys()
		if err != nil {
			panic(err)
		}
	}

	return &APIService{service: &service{store, privKey, pubKey, cache, c, l, validator.New()}}
}
