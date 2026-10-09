// Package devicekey reads the key pair a device enrols with: the agent's SSH host key, whose
// public half the device record stores as PEM.
package devicekey

import (
	"crypto/x509"
	"encoding/pem"
	"errors"

	gossh "golang.org/x/crypto/ssh"
)

// ErrInvalidPublicKey is returned when a device's stored key is not a PEM public key.
var ErrInvalidPublicKey = errors.New("invalid device public key")

// ParsePublicKey reads a device public key stored as PEM: an "RSA PUBLIC KEY" block as PKCS#1,
// which is what the agent sends, and any other block as PKIX. It returns ErrInvalidPublicKey
// when there is no PEM block, the block does not parse that way, or the key is of a type
// x/crypto/ssh cannot represent.
func ParsePublicKey(pemKey string) (gossh.PublicKey, error) { //nolint:ireturn // gossh.PublicKey is how x/crypto/ssh represents every key type
	block, _ := pem.Decode([]byte(pemKey))
	if block == nil {
		return nil, ErrInvalidPublicKey
	}

	var (
		key any
		err error
	)

	switch block.Type {
	case "RSA PUBLIC KEY":
		key, err = x509.ParsePKCS1PublicKey(block.Bytes)
	default:
		key, err = x509.ParsePKIXPublicKey(block.Bytes)
	}

	if err != nil {
		return nil, errors.Join(ErrInvalidPublicKey, err)
	}

	pub, err := gossh.NewPublicKey(key)
	if err != nil {
		return nil, errors.Join(ErrInvalidPublicKey, err)
	}

	return pub, nil
}
