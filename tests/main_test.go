package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"testing"

	log "github.com/sirupsen/logrus"
)

// keygen generates private and public keys required to startup a ShellHub instance.
func keygen() error {
	const PrivateKeyPermission uint = 0o600
	const PublicKeyPermission uint = 0o644

	const APIPrivatKeyPath string = "../api_private_key"
	const APIPublicKeyPath string = "../api_public_key"
	const SSHPrivateKey string = "../ssh_private_key"

	if _, err := os.Stat(SSHPrivateKey); os.IsNotExist(err) {
		sshPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return err
		}

		bytesSSHPrivateKey, err := x509.MarshalPKCS8PrivateKey(sshPrivateKey)
		if err != nil {
			return err
		}

		if err := os.WriteFile(SSHPrivateKey, pem.EncodeToMemory(
			&pem.Block{
				Type:  "PRIVATE KEY",
				Bytes: bytesSSHPrivateKey,
			},
		), os.FileMode(PrivateKeyPermission)); err != nil {
			return err
		}
	}

	if _, err := os.Stat(APIPrivatKeyPath); os.IsNotExist(err) {
		apiPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return err
		}

		bytesAPIPrivateKey, err := x509.MarshalPKCS8PrivateKey(apiPrivateKey)
		if err != nil {
			return err
		}

		if err := os.WriteFile(APIPrivatKeyPath, pem.EncodeToMemory(
			&pem.Block{
				Type:  "PRIVATE KEY",
				Bytes: bytesAPIPrivateKey,
			},
		), os.FileMode(PrivateKeyPermission)); err != nil {
			return err
		}

		bytesAPIPublicKey, err := x509.MarshalPKIXPublicKey(&apiPrivateKey.PublicKey)
		if err != nil {
			return err
		}

		if err := os.WriteFile(APIPublicKeyPath, pem.EncodeToMemory(
			&pem.Block{
				Type:  "PUBLIC KEY",
				Bytes: bytesAPIPublicKey,
			},
		), os.FileMode(PublicKeyPermission)); err != nil {
			return err
		}
	}

	return nil
}

func TestMain(m *testing.M) {
	// FIXME: Due to issue related on testcontainers-go, we are disabling Ryuk it as a temporary solution.
	//
	// https://github.com/testcontainers/testcontainers-go/issues/2445
	os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")

	if err := keygen(); err != nil {
		log.WithError(err).Error("failed to generate the ShellHub keys")

		os.Exit(1)
	}

	os.Exit(m.Run())
}
