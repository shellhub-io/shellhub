package yescrypt

import (
	yescrypt "github.com/openwall/yescrypt-go"
	log "github.com/sirupsen/logrus"
)

// Verify verifies a yescrypt hash against a given key.
func Verify(password, hash string) bool {
	hashed, err := yescrypt.Hash([]byte(password), []byte(hash))
	if err != nil {
		log.WithError(err).Debug("failed to hash the password for comparison")

		return false
	}

	return hash == string(hashed)
}
