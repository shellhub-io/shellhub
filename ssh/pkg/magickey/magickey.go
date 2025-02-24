package magickey

import (
	"crypto/rand"
	"crypto/rsa"
	"sync"

	log "github.com/sirupsen/logrus"
)

// GetRerefence gets a reference for a [*rsa.PrivateKey] used on ShellHub's SSH service. This reference is generate once
// and them used on every subsequence call.
var GetRerefence = sync.OnceValue(func() *rsa.PrivateKey {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.WithError(err).Fatal()
	}

	return key
})
