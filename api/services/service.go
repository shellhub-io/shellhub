package services

import (
	"crypto/rsa"

	"github.com/shellhub-io/shellhub/api/cache"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/geoip"
)

type service struct {
	store   store.Store
	privKey *rsa.PrivateKey
	pubKey  *rsa.PublicKey
	cache   cache.Cache
	client  interface{}
	locator geoip.Locator
}

type Service interface {
	DeviceService
	DeviceTags
	UserService
	SSHKeysService
	SessionService
	NamespaceService
	AuthService
	StatsService
}

func NewService(store store.Store, privKey *rsa.PrivateKey, pubKey *rsa.PublicKey, cache cache.Cache, c interface{}, l geoip.Locator) Service {
	if privKey == nil || pubKey == nil {
		var err error
		privKey, pubKey, err = LoadKeys()
		if err != nil {
			panic(err)
		}
	}

	return &service{store, privKey, pubKey, cache, c, l}
}
