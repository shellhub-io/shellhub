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
	// PublicKey is the OpenSSH authorized-keys blob of the key used for the
	// direct connection bridge (/ws/connect). The target host is external, so
	// its public key is not registered in ShellHub and must be supplied by the
	// browser. The matching private key never leaves the browser: the server
	// proxies each signing challenge over the websocket (see [Signer]).
	PublicKey string `json:"public_key"`
	// Host and Port are set only for the direct connection bridge (/ws/connect),
	// where the server dials the target SSH endpoint directly instead of routing
	// to a device over the reverse tunnel.
	Host string `json:"host"`
	Port int    `json:"port"`
	// KnownHostKey is the host key the browser confirmed (authorized_keys format)
	// for this external target. The server verifies the live host key against it
	// and aborts on mismatch. Empty means no verified key (legacy/uninitialized).
	KnownHostKey string `json:"known_host_key"`
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

func (c *Credentials) isPublicKey() bool {
	return c.Fingerprint != ""
}

// isPassword checks if connection is using password method.
func (c *Credentials) isPassword() bool {
	return !c.isPublicKey()
}

// Dimensions represents a web SSH terminal dimensions.
type Dimensions struct {
	Cols uint32 `json:"cols"`
	Rows uint32 `json:"rows"`
}

type Info struct {
	IP string `json:"ip"`
}
