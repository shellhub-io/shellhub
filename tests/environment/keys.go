package environment

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
)

// GenerateKeys creates the RSA key files ShellHub needs (api_private_key, api_public_key,
// ssh_private_key) under dir, skipping any that already exist.
func GenerateKeys(dir string) error {
	const privateKeyPerm = 0o600
	const publicKeyPerm = 0o644

	sshPrivateKeyPath := filepath.Join(dir, "ssh_private_key")
	apiPrivateKeyPath := filepath.Join(dir, "api_private_key")
	apiPublicKeyPath := filepath.Join(dir, "api_public_key")

	if _, err := os.Stat(sshPrivateKeyPath); os.IsNotExist(err) {
		key, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return err
		}

		der, err := x509.MarshalPKCS8PrivateKey(key)
		if err != nil {
			return err
		}

		if err := os.WriteFile(sshPrivateKeyPath, pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der}), os.FileMode(privateKeyPerm)); err != nil {
			return err
		}
	}

	if _, err := os.Stat(apiPrivateKeyPath); os.IsNotExist(err) {
		key, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return err
		}

		der, err := x509.MarshalPKCS8PrivateKey(key)
		if err != nil {
			return err
		}

		if err := os.WriteFile(apiPrivateKeyPath, pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der}), os.FileMode(privateKeyPerm)); err != nil {
			return err
		}

		pub, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
		if err != nil {
			return err
		}

		if err := os.WriteFile(apiPublicKeyPath, pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pub}), os.FileMode(publicKeyPerm)); err != nil {
			return err
		}
	}

	return nil
}
