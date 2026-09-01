package web

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"

	"golang.org/x/crypto/ssh"
)

// Credentials is what the browser supplies to log in to a device. Exactly one of the password
// and the key is used, depending on the method the device accepts.
type Credentials struct {
	// Device is the device what the session is open.
	Device string `json:"device"`
	// Username is the username in the device's OS.
	Username string `json:"username"`
	// Password is the password in the device's OS.
	Password string `json:"password"`
	// Fingerprint is the identifier of the public key used in the device's OS.
	Fingerprint string `json:"fingerprint"`
	// PublicKey is the browser's own enrolled SSH public key (an "ssh-ed25519 …"
	// line) in identity access mode. When set, the web terminal authenticates by
	// signing the SSH challenge with the matching non-extractable browser key,
	// and the gateway resolves it to the identity — no device credential and no
	// legacy public-key ACL. Empty in legacy and keyless-fallback sessions.
	PublicKey string `json:"public_key"`
	// UserID is the ShellHub account driving the web terminal, taken from the
	// X-ID header the gateway injects after authenticating the request (never
	// from the request body). It is the identity used in the identity access
	// mode, where no device credential is presented; empty in legacy mode.
	UserID string `json:"-"`
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

func (c *Credentials) isPassword() bool {
	return !c.isPublicKey()
}

// Dimensions represents a web SSH terminal dimensions.
//
// The sizes are 16 bits wide because that is what a pty can allocate; the width also makes
// the conversion to the int the pty request takes exact on any platform.
type Dimensions struct {
	Cols uint16 `json:"cols"`
	Rows uint16 `json:"rows"`
}

// Info is the client-supplied context recorded against the session — its originating IP,
// which the browser cannot be trusted for and which the handler overrides.
type Info struct {
	IP string `json:"ip"`
}

// readHostKey loads the public half of the SSH server's host key, which the bridge pins its
// loopback connection to.
func readHostKey(path string) (ssh.PublicKey, error) {
	pem, err := os.ReadFile(path) //nolint:gosec // path comes from the server's own configuration, not user input.
	if err != nil {
		return nil, errors.Join(ErrBridgeReadHostKey, err)
	}

	signer, err := ssh.ParsePrivateKey(pem)
	if err != nil {
		return nil, errors.Join(ErrBridgeReadHostKey, err)
	}

	return signer.PublicKey(), nil
}
