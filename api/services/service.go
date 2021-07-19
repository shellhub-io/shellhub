package services

import (
	"crypto/rsa"

	"github.com/shellhub-io/shellhub/api/store"
)

type service struct {
	store   store.Store
	privKey *rsa.PrivateKey
	pubKey  *rsa.PublicKey
}

type Service interface {
	DeviceService
	UserService
	SSHKeysService
	SessionService
	NamespaceService
	AuthService
	StatsService
}

func NewService(store store.Store, privKey *rsa.PrivateKey, pubKey *rsa.PublicKey) Service {
	if privKey == nil || pubKey == nil {
		var err error
		privKey, pubKey, err = LoadKeys()
		if err != nil {
			panic(err)
		}
	}

	return &service{store, privKey, pubKey}
}
