package web

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"errors"
)

type Credentials struct {
	// Device is the device what the session is open.
	Device string `json:"device"`
	// Username is the username in the device's OS.
	Username string `json:"username"`
	// Password is the password in the device's OS.
	Password string `json:"password"`
	// Fingerprint is the identifier of the public key used in the device's OS.
	Fingerprint string `json:"fingerprint"`
	Signature   string `json:"signature"`
}

func (c *Credentials) encryptPassword(key *rsa.PrivateKey) error {
	if c.Password == "" {
		return ErrCreditialsNoPassword
	}

	signed, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, &key.PublicKey, []byte(c.Password), nil)
	if err != nil {
		return errors.New("failed to sign the session's password")
	}

	c.Password = hex.EncodeToString(signed)

	return nil
}

func (c *Credentials) decryptPassword(key *rsa.PrivateKey) error {
	if c.Password == "" {
		return ErrCreditialsNoPassword
	}

	decoded, err := hex.DecodeString(c.Password)
	if err != nil {
		return errors.New("failed to decode the session's password")
	}

	decrypted, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, key, decoded, nil)
	if err != nil {
		return errors.New("failed to decrypt the session's password")
	}

	c.Password = string(decrypted)

	return nil
}

func (c *Credentials) isPublicKey() bool { // nolint: unused
	return c.Fingerprint != "" && c.Signature != ""
}

// isPassword checks if connection is using password method.
func (c *Credentials) isPassword() bool {
	return c.Password != ""
}

// Dimensions represents a web SSH terminal dimensions.
type Dimensions struct {
	Cols int `json:"cols"`
	Rows int `json:"rows"`
}

type Info struct {
	IP string `json:"ip"`
}
