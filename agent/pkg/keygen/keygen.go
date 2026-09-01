package keygen

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

var (
	// ErrPemDecode is returned when PEM decoding fails.
	ErrPemDecode = errors.New("PEM decode error")

	// ErrPathTraversal is returned when a filename escapes the declared base directory.
	ErrPathTraversal = errors.New("path escapes base directory")
)

// GeneratePrivateKey writes a new 2048-bit RSA key to filename in PEM form, creating the
// directory if needed. The file is written readable only by its owner.
func GeneratePrivateKey(filename string) error {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	_, err = os.Stat(filepath.Dir(filename))
	if os.IsNotExist(err) {
		if err = os.MkdirAll(filepath.Dir(filename), 0o700); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	f, err := os.Create(filename) //nolint:gosec // filename is a configured key path, not user-supplied taint input.
	if err != nil {
		return err
	}

	defer f.Close() //nolint:errcheck

	privateKey := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}

	err = pem.Encode(f, privateKey)
	if err != nil {
		return err
	}

	return f.Sync()
}

// ReadPublicKey loads a PEM-encoded private key from filename and returns its public half.
// It reports [ErrPemDecode] when the file holds no PEM block.
func ReadPublicKey(filename string) (*rsa.PublicKey, error) {
	data, err := os.ReadFile(filename) //nolint:gosec // filename is a configured key path, not user-supplied taint input.
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, ErrPemDecode
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return &key.PublicKey, nil
}

// EncodePublicKeyToPem renders key in the PEM form the server expects at enrolment.
func EncodePublicKeyToPem(key *rsa.PublicKey) []byte {
	return pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: x509.MarshalPKCS1PublicKey(key),
	})
}
