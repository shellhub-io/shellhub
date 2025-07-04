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

//go:generate mockery --name=Ticker --filename=ticker.go
type Ticker interface {
	// Init creates a new [time.Ticker] internally with [time.Duration] defined.
	Init(context.Context, time.Duration)
	// Tick waits for a ticker's tick and return the value. If ticker wasn't initialized, a [time.Time] with zero-value
	// will be returned.
	Tick() chan time.Time
	// Stop stops the ticker initialized. If ticker wasn't initialized, nothing happens.
	Stop()
}

type ticker struct {
	ticker *time.Ticker
	tick   chan time.Time
}

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

func (t *ticker) Tick() chan time.Time {
	return t.tick
}

func (t *ticker) Stop() {
	if t.ticker == nil {
		return
	}

	t.ticker.Stop()
}

// DNSProvider represents a DNS provider to generate certificates.
type DNSProvider string

// DigitalOceanDNSProvider represents the Digital Ocean DNS provider.
const DigitalOceanDNSProvider = "digitalocean"

type Config struct {
	// RootDir is the root directory for CertBot configurations.
	RootDir string
	// Staging defines if the CertBot will use the staging server to generate certificates.
	Staging bool
	// RenewedCallback is a callback called after certificate renew.
	RenewedCallback func()
}

type Certificate interface {
	String() string
	Generate(staging bool) error
}

type DefaultCertificate struct {
	RootDir string
	Domain  string

	ex Executor
	fs afero.Fs
}

func NewDefaultCertificate(domain string) Certificate {
	return &DefaultCertificate{
		Domain: domain,

		ex: NewExecutor(),
		fs: afero.NewOsFs(),
	}
}

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
func (d *DefaultCertificate) stopACMEServer(server *http.Server) {
	if err := server.Close(); err != nil {
		log.WithError(err).Fatal("could not stop ACME server")
	}
}

func (d *DefaultCertificate) Generate(staging bool) error {
	log.Info("generating SSL certificate")

	challengeDir := fmt.Sprintf("%s/.well-known/acme-challenge", os.TempDir())
	if err := d.fs.MkdirAll(challengeDir, 0o755); err != nil {
		log.WithError(err).Error("failed to create acme challenge on filesystem")

		return err
	}

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

	d.stopACMEServer(acmeServer)

	log.Info("generate run")

	return nil
}

func (d *DefaultCertificate) String() string {
	return d.Domain
}

type TunnelsCertificate struct {
	// Domain is the default domain used to generate certificate for Tunnels.
	Domain string
	// Provider is the DNS provider used to generate wildcard certificates.
	Provider DNSProvider
	// Token is a DNS token used to generate wildcard certificates.
	Token string

	ex Executor
	fs afero.Fs
}

func NewTunnelsCertificate(domain string, provider DNSProvider, token string) Certificate {
	return &TunnelsCertificate{
		Domain: domain,

		Provider: provider,
		Token:    token,

		ex: NewExecutor(),
		fs: afero.NewOsFs(),
	}
}

func (d *TunnelsCertificate) generateProviderCredentialsFile() (afero.File, error) {
	token := fmt.Sprintf("dns_%s_token = %s", d.Provider, d.Token)

	file, err := d.fs.Create(fmt.Sprintf("/etc/shellhub-gateway/%s.ini", string(d.Provider)))
	if err != nil {
		log.WithError(err).Error("failed to create shellhub-gateway file with dns provider token")

		return nil, err
	}

	if _, err := file.Write([]byte(token)); err != nil {
		log.WithError(err).Error("failed to write the token into credentials file")

		return nil, err
	}

	return file, nil
}

func (d *TunnelsCertificate) Generate(staging bool) error {
	log.Info("generating SSL certificate with DNS")

	file, err := d.generateProviderCredentialsFile()
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

func (d *TunnelsCertificate) String() string {
	return d.Domain
}

// CertBot handles the generation and renewal of SSL certificates.
type CertBot struct {
	Config *Config

	Certificates []Certificate

	ex Executor
	tk Ticker
	fs afero.Fs
}

func newCertBot(config *Config) *CertBot {
	return &CertBot{
		Config: config,

		ex: new(executor),
		tk: new(ticker),
		fs: afero.NewOsFs(),
	}
}

// ensureCertificates checks if the SSL certificate exists and generates if it doesn't.
func (cb *CertBot) ensureCertificates() {
	for _, certificate := range cb.Certificates {
		certPath := fmt.Sprintf("%s/live/%s/fullchain.pem", cb.Config.RootDir, certificate)
		if _, err := cb.fs.Stat(certPath); os.IsNotExist(err) {
			certificate.Generate(cb.Config.Staging)
		}
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
