package keygen

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

var ErrPemDecode = errors.New("PEM decode error")

func GeneratePrivateKey(filename string) error {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	_, err = os.Stat(filepath.Dir(filename))
	if os.IsNotExist(err) {
		// Create parent directory if it does not exist
		if err = os.MkdirAll(filepath.Dir(filename), 0o700); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	f, err := os.Create(filename)
	if err != nil {
		return err
	}

	defer f.Close()

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

func ReadPublicKey(filename string) (*rsa.PublicKey, error) {
	data, err := ioutil.ReadFile(filename)
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

func EncodePublicKeyToPem(key *rsa.PublicKey) []byte {
	return pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: x509.MarshalPKCS1PublicKey(key),
	})
}
