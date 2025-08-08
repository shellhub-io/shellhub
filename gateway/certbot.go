// Package main provides SSL certificate management functionality using CertBot.
// It supports both HTTP-01 and DNS-01 challenge types for certificate generation
// and automatic renewal of SSL certificates.
package main

import (
	"context"
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

// Executor provides an interface for executing system commands.
// This interface allows for easy mocking in tests and provides
// a clean abstraction over the exec package.
//
//go:generate mockery --name=Executor --filename=executor.go
type Executor interface {
	// Command creates a new *exec.Cmd with the given name and arguments.
	Command(name string, arg ...string) *exec.Cmd
	// Run executes the given command and waits for it to complete.
	Run(cmd *exec.Cmd) error
}

// executor is the default implementation of the Executor interface.
type executor struct{}

// NewExecutor creates a new Executor instance.
func NewExecutor() Executor {
	return &executor{}
}

// Command creates a new *exec.Cmd with the given name and arguments.
func (e *executor) Command(name string, arg ...string) *exec.Cmd {
	return exec.Command(name, arg...)
}

// Run executes the given command and waits for it to complete.
func (e *executor) Run(cmd *exec.Cmd) error {
	return cmd.Run()
}

// Ticker provides an interface for time-based operations with context support.
// This interface allows for easy mocking in tests and provides a clean
// abstraction over the time package's ticker functionality.
//
//go:generate mockery --name=Ticker --filename=ticker.go
type Ticker interface {
	// Init creates a new time.Ticker internally with the specified duration.
	// The ticker will respect the provided context for cancellation.
	Init(context.Context, time.Duration)
	// Tick returns a channel that receives the current time on each tick.
	// If the ticker wasn't initialized, the channel will be nil.
	Tick() chan time.Time
	// Stop stops the ticker. If the ticker wasn't initialized, this is a no-op.
	Stop()
}

// ticker is the default implementation of the Ticker interface.
type ticker struct {
	ticker *time.Ticker
	tick   chan time.Time
}

// Init creates a new time.Ticker internally with the specified duration.
// It starts a goroutine that forwards ticker events to the tick channel
// and handles context cancellation.
func (t *ticker) Init(ctx context.Context, duration time.Duration) {
	t.ticker = time.NewTicker(duration)
	t.tick = make(chan time.Time)

	go func() {
		defer close(t.tick)

		for {
			select {
			case <-ctx.Done():
				return
			case ticked, ok := <-t.ticker.C:
				if !ok {
					return
				}

				t.tick <- ticked
			}
		}
	}()
}

// Tick returns a channel that receives the current time on each tick.
func (t *ticker) Tick() chan time.Time {
	return t.tick
}

// Stop stops the ticker. If the ticker wasn't initialized, this is a no-op.
func (t *ticker) Stop() {
	if t.ticker == nil {
		return
	}

	t.ticker.Stop()
}

// DNSProvider represents a DNS provider that can be used for DNS-01 challenges
// when generating SSL certificates.
type DNSProvider string

// DigitalOceanDNSProvider represents the Digital Ocean DNS provider.
const DigitalOceanDNSProvider = "digitalocean"

// CloudflareDNSProvider represents the Cloudflare DNS provider.
const CloudflareDNSProvider = "cloudflare"

// Config holds the configuration for CertBot operations.
type Config struct {
	// RootDir is the root directory where CertBot stores its configurations
	// and generated certificates. Typically "/etc/letsencrypt".
	RootDir string
	// Staging defines whether CertBot should use Let's Encrypt's staging server
	// instead of the production server. Useful for testing to avoid rate limits.
	Staging bool
	// RenewedCallback is an optional callback function that gets called
	// after a certificate is successfully renewed.
	RenewedCallback func()
}

// Certificate represents an SSL certificate that can be generated using CertBot.
type Certificate interface {
	// String returns a string representation of the certificate, typically the domain name.
	String() string
	// Check checks if the environment is ready for certificate generation or renewal.
	Check() error
	// Generate creates the SSL certificate using CertBot.
	// The staging parameter determines whether to use Let's Encrypt's staging server.
	Generate(staging bool) error
}

// DefaultCertificate represents a standard SSL certificate that uses HTTP-01 challenge
// for domain validation. This is suitable for single domains where you have control
// over the web server.
type DefaultCertificate struct {
	// RootDir is the root directory for certificate storage.
	RootDir string
	// Domain is the domain name for which the certificate will be generated.
	Domain string

	ex Executor
	fs afero.Fs
}

// NewDefaultCertificate creates a new DefaultCertificate instance for the given domain.
func NewDefaultCertificate(rootdir string, domain string) Certificate {
	return &DefaultCertificate{
		RootDir: rootdir,
		Domain:  domain,

		ex: NewExecutor(),
		fs: afero.NewOsFs(),
	}
}

// startACMEServer starts a local HTTP server on port 80 to handle ACME HTTP-01 challenges.
// This server serves files from the .well-known/acme-challenge directory which is
// required for Let's Encrypt domain validation.
func (d *DefaultCertificate) startACMEServer() *http.Server {
	mux := http.NewServeMux()
	mux.Handle(
		"/.well-known/acme-challenge/",
		http.StripPrefix(
			"/.well-known/acme-challenge/",
			http.FileServer(
				http.Dir(filepath.Join(d.RootDir, ".well-known/acme-challenge")),
			),
		),
	)

	server := &http.Server{ //nolint:gosec
		Handler: mux,
	}

	listener, err := net.Listen("tcp", ":80") //nolint:gosec
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

// stopACMEServer gracefully stops the local ACME HTTP server.
func (d *DefaultCertificate) stopACMEServer(server *http.Server) {
	if err := server.Close(); err != nil {
		log.WithError(err).Fatal("could not stop ACME server")
	}
}

func (d *DefaultCertificate) Check() error {
	if d.Domain == "" {
		return errors.New("domain is required for certificate generation")
	}

	if d.RootDir == "" {
		return errors.New("root directory is required for certificate generation")
	}

	if _, err := d.fs.Stat(d.RootDir); os.IsNotExist(err) {
		if err := d.fs.MkdirAll(d.RootDir, 0o755); err != nil {
			log.WithError(err).Error("failed to create root directory for certificate generation")

			return err
		}
	}

	return nil
}

// Generate creates an SSL certificate for the domain using HTTP-01 challenge.
// It starts a local HTTP server to handle the ACME challenge, runs CertBot,
// and then stops the server.
func (d *DefaultCertificate) Generate(staging bool) error {
	log.Info("generating SSL certificate")

	// Create the ACME challenge directory
	challengeDir := fmt.Sprintf("%s/.well-known/acme-challenge", os.TempDir())
	if err := d.fs.MkdirAll(challengeDir, 0o755); err != nil {
		log.WithError(err).Error("failed to create acme challenge on filesystem")

		return err
	}

	// Start the ACME server to handle HTTP-01 challenges
	acmeServer := d.startACMEServer()

	args := []string{
		"certonly",
		"--non-interactive",
		"--agree-tos",
		"--register-unsafely-without-email",
		"--webroot",
		"--webroot-path", d.RootDir,
		"--preferred-challenges", "http",
		"-n",
		"-d",
		d.Domain,
	}

	if staging {
		log.Info("running generate with staging")

		args = append(args, "--staging")
	}

	// Build the CertBot command
	cmd := d.ex.Command(
		"certbot",
		args...,
	)

	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr

	if err := d.ex.Run(cmd); err != nil {
		log.Error("Failed to generate SSL certificate")

		return err
	}

	// Stop the ACME server
	d.stopACMEServer(acmeServer)

	log.Info("generate run")

	return nil
}

// String returns the domain name as the string representation of the certificate.
func (d *DefaultCertificate) String() string {
	return d.Domain
}

// WebEndpointsCertificate represents a wildcard SSL certificate that uses DNS-01 challenge
// for domain validation. This is suitable for wildcard certificates (*.example.com)
// where you have control over the DNS records.
type WebEndpointsCertificate struct {
	// Domain is the base domain used to generate wildcard certificates.
	Domain string
	// Provider is the DNS provider used for DNS-01 challenges.
	Provider DNSProvider
	// Token is the API token for the DNS provider.
	Token string

	ex Executor
	fs afero.Fs
}

// NewWebEndpointsCertificate creates a new TunnelsCertificate instance for generating
// wildcard certificates using DNS-01 challenges.
func NewWebEndpointsCertificate(domain string, provider DNSProvider, token string) Certificate {
	return &WebEndpointsCertificate{
		Domain: domain,

		Provider: provider,
		Token:    token,

		ex: NewExecutor(),
		fs: afero.NewOsFs(),
	}
}

// generateProviderCredentialsFile creates a credentials file for the DNS provider.
// This file contains the API token needed for DNS-01 challenges.
func (d *WebEndpointsCertificate) generateProviderCredentialsFile() (afero.File, error) {
	tokenLine := fmt.Sprintf("dns_%s_token = %s", d.Provider, d.Token)

	// Certbot Cloudflare plugin expects dns_cloudflare_api_token
	if d.Provider == CloudflareDNSProvider {
		tokenLine = fmt.Sprintf("dns_cloudflare_api_token = %s", d.Token)
	}

	file, err := d.fs.Create(fmt.Sprintf("/etc/shellhub-gateway/%s.ini", string(d.Provider)))
	if err != nil {
		log.WithError(err).Error("failed to create shellhub-gateway file with dns provider token")

		return nil, err
	}

	if _, err := file.Write([]byte(tokenLine)); err != nil {
		log.WithError(err).Error("failed to write the token into credentials file")

		return nil, err
	}

	return file, nil
}

func (d *WebEndpointsCertificate) Check() error {
	if d.Domain == "" {
		return errors.New("domain is required for certificate generation")
	}

	if d.Provider == "" {
		return errors.New("DNS provider is required for certificate generation")
	}

	if d.Token == "" {
		return errors.New("DNS provider token is required for certificate generation")
	}

	if _, err := d.fs.Stat("/etc/shellhub-gateway"); os.IsNotExist(err) {
		if err := d.fs.MkdirAll("/etc/shellhub-gateway", 0o755); err != nil {
			log.WithError(err).Error("failed to create /etc/shellhub-gateway directory")

			return err
		}
	}

	if _, err := d.generateProviderCredentialsFile(); err != nil {
		log.WithError(err).Error("failed to generate provider credentials file")

		return err
	}

	return nil
}

// Generate creates a wildcard SSL certificate for the domain using DNS-01 challenge.
// It creates a credentials file for the DNS provider, runs CertBot with DNS plugin,
// and generates a wildcard certificate.
func (d *WebEndpointsCertificate) Generate(staging bool) error {
	log.Info("generating SSL certificate with DNS")

	// Create the DNS provider credentials file
	file, err := d.generateProviderCredentialsFile()
	if err != nil {
		log.WithError(err).Error("failed to generate INI file")

		return err
	}

	// Build the CertBot command arguments for DNS-01 challenge
	args := []string{
		"certonly",
		"--non-interactive",
		"--agree-tos",
		"--register-unsafely-without-email",
		"--cert-name",
		fmt.Sprintf("*.%s", d.Domain),
		fmt.Sprintf("--dns-%s", d.Provider),
		fmt.Sprintf("--dns-%s-credentials", d.Provider),
		file.Name(),
		"-d",
		fmt.Sprintf("*.%s", d.Domain),
	}

	if staging {
		log.Info("running generate with staging on dns")

		args = append(args, "--staging")
	}

	cmd := d.ex.Command( //nolint:gosec
		"certbot",
		args...,
	)

	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr

	if err := d.ex.Run(cmd); err != nil {
		log.WithError(err).Error("failed to generate SSL certificate")

		return err
	}

	log.Info("generate run on dns")

	return nil
}

// String returns the domain name as the string representation of the certificate.
func (d *WebEndpointsCertificate) String() string {
	return d.Domain
}

// CertBot is the main structure that handles SSL certificate generation and renewal.
// It manages multiple certificates and provides automatic renewal functionality.
type CertBot struct {
	// Config holds the configuration for CertBot operations.
	Config *Config

	// Certificates is a list of certificates to manage.
	Certificates []Certificate

	ex Executor
	tk Ticker
	fs afero.Fs
}

// newCertBot creates a new CertBot instance with the given configuration.
func newCertBot(config *Config) *CertBot {
	return &CertBot{
		Config: config,

		ex: new(executor),
		tk: new(ticker),
		fs: afero.NewOsFs(),
	}
}

// ensureCertificates checks if SSL certificates exist for all managed domains.
// If a certificate doesn't exist, it generates a new one.
func (cb *CertBot) ensureCertificates() {
	for _, certificate := range cb.Certificates {
		certPath := fmt.Sprintf("%s/live/%s/fullchain.pem", cb.Config.RootDir, certificate)
		if _, err := cb.fs.Stat(certPath); os.IsNotExist(err) {
			certificate.Generate(cb.Config.Staging)
		}
	}
}

// executeRenewCertificates runs the CertBot renew command to check and renew
// certificates that are close to expiration.
func (cb *CertBot) executeRenewCertificates() error {
	args := []string{
		"renew",
	}

	if cb.Config.Staging {
		log.Info("running renew with staging")

		args = append(args, "--staging")
	}

	for _, certificate := range cb.Certificates {
		if err := certificate.Check(); err != nil {
			log.WithError(err).Error("certificate check failed")

			return err
		}
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

// renewCertificates starts a background process that periodically checks and renews
// SSL certificates. It runs in a loop with the specified duration between checks.
// The process respects context cancellation for graceful shutdown.
func (cb *CertBot) renewCertificates(ctx context.Context, duration time.Duration) {
	log.Info("starting SSL certificate renewal process")

	cb.tk.Init(ctx, duration)
	defer cb.tk.Stop()

	ticker := cb.tk.Tick()

	for {
		select {
		case <-ctx.Done():
			log.Info("renew certificates loop was closed due context cancellation")

			return
		case <-ticker:
			log.Info("checking if SSL certificate needs to be renewed")
			if err := cb.executeRenewCertificates(); err != nil {
				log.WithError(err).Error("failed to renew SSL certificate")

				continue
			}

			log.Info("ssl certificate successfully renewed")
			cb.Config.RenewedCallback()
		}
	}
}
