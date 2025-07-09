// Package magickey provides RSA key generation for ShellHub SSH service.
package magickey

import (
	"crypto/rand"
	"crypto/rsa"
	"sync"

	log "github.com/sirupsen/logrus"
)

// GetReference returns a singleton RSA private key for ShellHub SSH service.
// The key is generated once and reused across all subsequent calls.
var GetReference = sync.OnceValue(func() *rsa.PrivateKey {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.WithError(err).Fatal()
	}

	return key
})
