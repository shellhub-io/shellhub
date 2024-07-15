package main

import (
	"fmt"
	"github.com/pkg/errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// CertBot handles the generation and renewal of SSL certificates.
type CertBot struct {
	rootDir         string
	domain          string
	staging         bool
	renewedCallback func()
}

// ensureCertificates checks if the SSL certificate exists and generates it if not.
func (cb *CertBot) ensureCertificates() {
	certPath := fmt.Sprintf("%s/live/%s/fullchain.pem", cb.rootDir, cb.domain)
	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		cb.generateCertificate()
	}
}

// generateCertificate generates a new SSL certificate using Certbot.
func (cb *CertBot) generateCertificate() {
	fmt.Println("Generating SSL certificate")

	challengeDir := fmt.Sprintf("%s/.well-known/acme-challenge", cb.rootDir)
	if err := os.MkdirAll(challengeDir, 0755); err != nil {
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

// renewCertificates periodically renews the SSL certificates.
func (cb *CertBot) renewCertificates() {
	for range time.Tick(24 * time.Hour) {
		fmt.Println("Checking if SSL certificate needs to be renewed")
		cmd := exec.Command(
			"certbot",
			"renew",
			"--staging",
			"--webroot",
			"--webroot-path",
			cb.rootDir,
		)
		if cb.staging {
			cmd.Args = append(cmd.Args, "--staging")
		}
		if err := cmd.Run(); err != nil {
			log.Fatal("Failed to renew SSL certificate")
		}
		fmt.Println("SSL certificate successfully renewed")
		cb.renewedCallback()
	}
}
