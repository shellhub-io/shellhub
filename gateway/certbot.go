package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
)

// DNSProvider represents a DNS provider to generate certificates.
type DNSProvider string

// DigitalOceanDNSProvider represents the Digital Ocean DNS provider.
const DigitalOceanDNSProvider = "digitalocean"

type tunnels struct {
	domain string
	token  string
}

// CertBot handles the generation and renewal of SSL certificates.
type CertBot struct {
	rootDir         string
	domain          string
	staging         bool
	renewedCallback func()
	tunnels         *tunnels
}

// ensureCertificates checks if the SSL certificate exists and generates it if not.
func (cb *CertBot) ensureCertificates() {
	certPath := fmt.Sprintf("%s/live/%s/fullchain.pem", cb.rootDir, cb.domain)
	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		cb.generateCertificate()
	}

	if cb.tunnels != nil {
		// NOTE: We are recreating the INI file every time to ensure it has the latest token from the environment.
		cb.generateProviderCredentialsFile(DigitalOceanDNSProvider)

		certPath := fmt.Sprintf("%s/live/*.%s/fullchain.pem", cb.rootDir, cb.tunnels.domain)
		if _, err := os.Stat(certPath); os.IsNotExist(err) {
			cb.generateCertificateFromDNS(DigitalOceanDNSProvider)
		}
	}
}

// generateCertificate generates a new SSL certificate using Certbot.
func (cb *CertBot) generateCertificate() {
	fmt.Println("Generating SSL certificate")

	challengeDir := fmt.Sprintf("%s/.well-known/acme-challenge", cb.rootDir)
	if err := os.MkdirAll(challengeDir, 0o755); err != nil {
		log.Fatal(err)
	}

	acmeServer := cb.startACMEServer()

	cmd := exec.Command(
		"certbot",
		"certonly",
		"--non-interactive",
		"--agree-tos",
		"--register-unsafely-without-email",
		"--webroot",
		"--webroot-path", cb.rootDir,
		"--preferred-challenges", "http",
		"-n",
		"-d",
		cb.domain,
	)
	if cb.staging {
		cmd.Args = append(cmd.Args, "--staging")
	}
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr

	if err := cmd.Run(); err != nil {
		log.Fatal("Failed to generate SSL certificate")
	}

	cb.stopACMEServer(acmeServer)
}

func (cb *CertBot) generateProviderCredentialsFile(provider DNSProvider) (*os.File, error) {
	token := fmt.Sprintf("dns_%s_token = %s", provider, cb.tunnels.token)
	file, err := os.Create(fmt.Sprintf("/etc/shellhub-gateway/%s.ini", string(provider)))
	if err != nil {
		return nil, err
	}

	file.Write([]byte(token))

	return file, nil
}

func (cb *CertBot) generateCertificateFromDNS(provider DNSProvider) {
	fmt.Println("Generating SSL certificate with DNS")

	file, err := cb.generateProviderCredentialsFile(provider)
	if err != nil {
		log.Fatalf("Failed to generate INI file: %v", err)
	}

	cmd := exec.Command( //nolint:gosec
		"certbot",
		"certonly",
		"--non-interactive",
		"--agree-tos",
		"--register-unsafely-without-email",
		"--cert-name",
		fmt.Sprintf("*.%s", cb.tunnels.domain),
		fmt.Sprintf("--dns-%s", provider),
		fmt.Sprintf("--dns-%s-credentials", provider),
		file.Name(),
		"-d",
		fmt.Sprintf("*.%s", cb.tunnels.domain),
	)
	if cb.staging {
		cmd.Args = append(cmd.Args, "--staging")
	}
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr

	if err := cmd.Run(); err != nil {
		log.Fatal("Failed to generate SSL certificate")
	}
}

// startACMEServer starts a local HTTP server for the ACME challenge.
func (cb *CertBot) startACMEServer() *http.Server {
	mux := http.NewServeMux()
	mux.Handle(
		"/.well-known/acme-challenge/",
		http.StripPrefix(
			"/.well-known/acme-challenge/",
			http.FileServer(
				http.Dir(filepath.Join(cb.rootDir, ".well-known/acme-challenge")),
			),
		),
	)

	server := &http.Server{
		Handler: mux,
	}

	listener, err := net.Listen("tcp", ":80")
	if err != nil {
		log.Fatalf("Failed to start ACME server listener: %v", err)
	}

	go func() {
		if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("ACME server error: %v", err)
		}
	}()

	return server
}

// stopACMEServer stops the local ACME server.
func (cb *CertBot) stopACMEServer(server *http.Server) {
	if err := server.Close(); err != nil {
		log.Fatalf("Could not stop ACME server: %v", err)
	}
}

func (cb *CertBot) executeRenewCertificates() error {
	cmd := exec.Command( //nolint:gosec
		"certbot",
		"renew",
	)

	if cb.staging {
		cmd.Args = append(cmd.Args, "--staging")
	}

	if err := cmd.Run(); err != nil {
		log.Println("Failed to renew SSL certificate")

		return err
	}

	return nil
}

// renewCertificates periodically renews the SSL certificates.
func (cb *CertBot) renewCertificates() {
	fmt.Println("Starting SSL certificate renewal process")

	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()
	for range ticker.C {
		fmt.Println("Checking if SSL certificate needs to be renewed")
		if err := cb.executeRenewCertificates(); err != nil {
			log.Fatal("Failed to renew SSL certificate")
		}

		fmt.Println("SSL certificate successfully renewed")
		cb.renewedCallback()
	}
}
