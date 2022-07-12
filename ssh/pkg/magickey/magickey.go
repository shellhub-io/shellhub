package magickey

import (
	"crypto/rand"
	"crypto/rsa"
	"sync"

	log "github.com/sirupsen/logrus"
)

var lock = &sync.Mutex{}

var magicKey *rsa.PrivateKey

func GetRerefence() *rsa.PrivateKey {
	if magicKey == nil {
		lock.Lock()
		defer lock.Unlock()

		key, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			log.WithError(err).Fatal()
		}

		magicKey = key
	}

	return magicKey
}
