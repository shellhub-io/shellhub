package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
)

type DeviceAttributes struct {
	ID         string `json:"id"`
	PrettyName string `json:"pretty_name"`
}

type AuthRequest struct {
	Identity   *DeviceIdentity   `json:"identity"`
	Attributes *DeviceAttributes `json:"attributes"`
	PublicKey  string            `json:"public_key"`
	TenantID   string            `json:"tenant_id"`
	Version    string            `json:"version"`
	Sessions   []string          `json:"sessions,omitempty"`
}

type AuthResponse struct {
	UID       string `json:"uid"`
	Token     string `json:"token"`
	Name      string `json:"name"`
	Namespace string `json: "namespace"`
}

func generatePrivateKey(filename string) error {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	f, err := os.Create(filename)
	if err != nil {
		return err
	}

	defer f.Close()

	var privateKey = &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}

	err = pem.Encode(f, privateKey)
	if err != nil {
		return err
	}

	return f.Sync()
}

func readPublicKey(filename string) (*rsa.PublicKey, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("Failed to decode PEM")
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return &key.PublicKey, nil
}
