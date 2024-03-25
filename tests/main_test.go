package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/testcontainers/testcontainers-go/modules/compose"
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
	ctx := context.Background()

	log.Info("Starting the image cache process")

	if err := keygen(); err != nil {
		log.WithError(err).Error("failed to generate the ShellHub keys")

		os.Exit(1)
	}

	// NOTICE: It is used to cache the images, avoiding the override when running in parallel.
	tcDc, err := compose.NewDockerCompose("../docker-compose.yml", "../docker-compose.test.yml")
	if err != nil {
		log.WithError(err).Error("failed to screate the ShellHub instance from docker cmpose files")

		os.Exit(1)
	}

	if err := tcDc.Up(ctx); err != nil {
		log.WithError(err).Error("failed to startup ShellHub instance")

		os.Exit(1)
	}

	if err := tcDc.Down(ctx, compose.RemoveOrphans(true), compose.RemoveVolumes(true)); err != nil {
		log.WithError(err).Error("failed to teardown ShellHub instance")

		os.Exit(1)
	}

	agent, err := NewAgentContainer(ctx, "")
	if err != nil {
		log.WithError(err).Error("failed to startup the Agent")

		os.Exit(1)
	}

	if err := agent.Stop(ctx, nil); err != nil {
		log.WithError(err).Error("failed to stop the Agent")

		os.Exit(1)
	}

	log.Info("Image cache process done")

	os.Exit(m.Run())
}
