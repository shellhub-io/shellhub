package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

//go:generate mockery --name=Executor --filename=executor.go
type Executor interface {
	Command(name string, arg ...string) *exec.Cmd
	Run(cmd *exec.Cmd) error
}

type executor struct{}

func NewExecutor() Executor {
	return &executor{}
}

func (e *executor) Command(name string, arg ...string) *exec.Cmd {
	return exec.Command(name, arg...)
}

func (e *executor) Run(cmd *exec.Cmd) error {
	return cmd.Run()
}

// DNSProvider represents a DNS provider to generate certificates.
type DNSProvider string

// DigitalOceanDNSProvider represents the Digital Ocean DNS provider.
const DigitalOceanDNSProvider = "digitalocean"

type Tunnels struct {
	// Domain is the default domain used to generate certificate for Tunnels.
	Domain string
	// Provider is the DNS provider used to generate wildcard certificates.
	Provider DNSProvider
	// Token is a DNS token used to generate wildcard certificates.
	Token string
}

type Config struct {
	// RootDir is the root directory for CertBot configurations.
	RootDir string
	// Domain is the default domain used to generate certificate for ShellHub.
	Domain string
	// Staging defines if the CertBot will use the staging server to generate certificates.
	Staging bool
	// RenewedCallback is a callback called after certificate renew.
	RenewedCallback func()

	Tunnels *Tunnels
}

// CertBot handles the generation and renewal of SSL certificates.
type CertBot struct {
	Config *Config

	ex Executor
	fs afero.Fs
}

func newCertBot(config *Config) *CertBot {
	return &CertBot{
		Config: config,

		ex: new(executor),
		fs: afero.NewOsFs(),
	}
}

// ensureCertificates checks if the SSL certificate exists and generates it if not.
func (cb *CertBot) ensureCertificates() {
	certPath := fmt.Sprintf("%s/live/%s/fullchain.pem", cb.Config.RootDir, cb.Config.Domain)
	if _, err := cb.fs.Stat(certPath); os.IsNotExist(err) {
		cb.generateCertificate()
	}

	if cb.Config.Tunnels != nil {
		// NOTE: We are recreating the INI file every time to ensure it has the latest token from the environment.
		cb.generateProviderCredentialsFile()

		certPath := fmt.Sprintf("%s/live/*.%s/fullchain.pem", cb.Config.RootDir, cb.Config.Tunnels.Domain)
		if _, err := cb.fs.Stat(certPath); os.IsNotExist(err) {
			if err := cb.generateCertificateFromDNS(); err != nil {
				log.WithError(err).Fatal("failed to generate the certificate from DNS")
			}
		}
	}
}

// generateCertificate generates a new SSL certificate using Certbot.
func (cb *CertBot) generateCertificate() {
	log.Info("generating SSL certificate")

	challengeDir := fmt.Sprintf("%s/.well-known/acme-challenge", cb.Config.RootDir)
	if err := cb.fs.MkdirAll(challengeDir, 0o755); err != nil {
		log.WithError(err).Fatal("failed to create acme challenge on filesystem")
	}

	acmeServer := cb.startACMEServer()

	cmd := cb.ex.Command(
		"certbot",
		"certonly",
		"--non-interactive",
		"--agree-tos",
		"--register-unsafely-without-email",
		"--webroot",
		"--webroot-path", cb.Config.RootDir,
		"--preferred-challenges", "http",
		"-n",
		"-d",
		cb.Config.Domain,
	)
	if cb.Config.Staging {
		log.Info("running generate with staging")

		cmd.Args = append(cmd.Args, "--staging")
	}
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr

	if err := cmd.Run(); err != nil {
		log.Fatal("Failed to generate SSL certificate")
	}

	cb.stopACMEServer(acmeServer)

	log.Info("generate run")
}

func (cb *CertBot) generateProviderCredentialsFile() (afero.File, error) {
	token := fmt.Sprintf("dns_%s_token = %s", cb.Config.Tunnels.Provider, cb.Config.Tunnels.Token)
	file, err := cb.fs.Create(fmt.Sprintf("/etc/shellhub-gateway/%s.ini", string(cb.Config.Tunnels.Provider)))
	if err != nil {
		log.WithError(err).Error("failed to create shellhub-gateway file with dns provider token")

		return nil, err
	}

	file.Write([]byte(token))

	return file, nil
}

func (cb *CertBot) generateCertificateFromDNS() error {
	log.Info("generating SSL certificate with DNS")

	file, err := cb.generateProviderCredentialsFile()
	if err != nil {
		log.WithError(err).Error("failed to generate INI file")

		return err
	}

	args := []string{
		"certonly",
		"--non-interactive",
		"--agree-tos",
		"--register-unsafely-without-email",
		"--cert-name",
		fmt.Sprintf("*.%s", cb.Config.Tunnels.Domain),
		fmt.Sprintf("--dns-%s", cb.Config.Tunnels.Provider),
		fmt.Sprintf("--dns-%s-credentials", cb.Config.Tunnels.Provider),
		file.Name(),
		"-d",
		fmt.Sprintf("*.%s", cb.Config.Tunnels.Domain),
	}

	if cb.Config.Staging {
		log.Info("running generate with staging on dns")

		args = append(args, "--staging")
	}

	cmd := cb.ex.Command( //nolint:gosec
		"certbot",
		args...,
	)

	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr

	if err := cb.ex.Run(cmd); err != nil {
		log.WithError(err).Error("failed to generate SSL certificate")

		return err
	}

	log.Info("generate run on dns")

	return nil
}

// startACMEServer starts a local HTTP server for the ACME challenge.
func (cb *CertBot) startACMEServer() *http.Server {
	mux := http.NewServeMux()
	mux.Handle(
		"/.well-known/acme-challenge/",
		http.StripPrefix(
			"/.well-known/acme-challenge/",
			http.FileServer(
				http.Dir(filepath.Join(cb.Config.RootDir, ".well-known/acme-challenge")),
			),
		),
	)

	server := &http.Server{
		Handler: mux,
	}

	listener, err := net.Listen("tcp", ":80")
	if err != nil {
		log.WithError(err).Fatal("failed to start ACME server listener")
	}

	go func() {
		if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.WithError(err).Fatal("acme server error")
		}
	}()

	return server
}

// stopACMEServer stops the local ACME server.
func (cb *CertBot) stopACMEServer(server *http.Server) {
	if err := server.Close(); err != nil {
		log.WithError(err).Fatal("could not stop ACME server")
	}
}

func (cb *CertBot) executeRenewCertificates() error {
	args := []string{
		"renew",
	}

	if cb.Config.Staging {
		log.Info("running renew with staging")

		args = append(args, "--staging")
	}

	cmd := cb.ex.Command( //nolint:gosec
		"certbot",
		args...,
	)

	if err := cb.ex.Run(cmd); err != nil {
		return err
	}

	log.Info("renew run")

	return nil
}

// renewCertificates periodically renews the SSL certificates.
func (cb *CertBot) renewCertificates() {
	log.Info("starting SSL certificate renewal process")

	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		log.Info("checking if SSL certificate needs to be renewed")
		if err := cb.executeRenewCertificates(); err != nil {
			log.WithError(err).Error("failed to renew SSL certificate")

			continue
		}

		log.Info("ssl certificate successfully renewed")
		cb.Config.RenewedCallback()
	}
}
