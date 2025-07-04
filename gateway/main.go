package main

import (
	"context"
	"time"

	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/loglevel"
	log "github.com/sirupsen/logrus"
)

const (
	// defaultNginxRootDir is the default base directory for Nginx configuration files.
	// This directory typically contains the main nginx.conf and site configuration files.
	defaultNginxRootDir = "/etc/nginx"

	// defaultNginxTemplateDir is the default directory where Nginx template files are stored.
	// Template files are used to generate dynamic Nginx configurations based on runtime settings.
	defaultNginxTemplateDir = "/templates"

	// defaultCertBotRootDir is the default directory where Certbot keeps
	// generated certificates, keys, and related assets.
	// This follows the standard Let's Encrypt directory structure.
	defaultCertBotRootDir = "/etc/letsencrypt"

	// defaultTickerRenewCertificates defines the interval for automatic certificate renewal checks.
	// Certificates are checked for renewal every 24 hours to ensure they remain valid.
	defaultTickerRenewCertificates = 24 * time.Hour
)

// Gateway represents the main gateway service that orchestrates Nginx configuration
// management and SSL certificate provisioning.
type Gateway struct {
	// Config holds the gateway's configuration settings including domain,
	// environment, and SSL settings.
	Config *GatewayConfig

	// Controller manages Nginx configuration generation, template processing,
	// and server lifecycle operations.
	Controller *NginxController

	// Certbot handles SSL certificate provisioning and renewal through Let's Encrypt.
	// This field is nil when SSL is not enabled.
	Certbot *CertBot
}

// NewGateway creates a new Gateway instance with the provided configuration and controller.
// The Certbot component is initially set to nil and will be initialized only when
// SSL is explicitly enabled through EnableSSL().
func NewGateway(config *GatewayConfig, controller *NginxController) *Gateway {
	return &Gateway{
		Config:     config,
		Controller: controller,
		Certbot:    nil,
	}
}

// EnableSSL initializes and configures SSL certificate management for the gateway.
// This method sets up Certbot with the gateway's domain configuration and establishes
// automatic certificate provisioning and renewal.
//
// The renewal callback is configured to reload Nginx when certificates are renewed,
// ensuring the server uses the latest certificates without manual intervention.
func (g *Gateway) EnableSSL() {
	g.Certbot = newCertBot(&Config{
		Domain:          g.Config.Domain,
		RootDir:         defaultCertBotRootDir,
		RenewedCallback: g.Controller.reload,
		Tunnels:         nil,
	})

	g.Certbot.ensureCertificates()
	g.Certbot.executeRenewCertificates()
}

func (g *Gateway) EnableTunnels() {
	domain := g.Config.Domain

	if g.Config.TunnelsDomain != "" {
		domain = g.Config.TunnelsDomain
	}

	g.Certbot.Config.Tunnels = &Tunnels{
		Domain:   domain,
		Provider: g.Config.TunnelsDNSProvider,
		Token:    g.Config.TunnelsDNSProviderToken,
	}
}

// Watch enables live monitoring of Nginx configuration template files.
//
// This method is typically used in development environments to automatically
// detect and apply configuration changes without requiring service restarts.
//
// The watching mechanism monitors the template directory for file changes
// and triggers configuration regeneration when modifications are detected.
func (g *Gateway) Watch() {
	go g.Controller.watchConfigTemplates()
}

// Start begins the gateway service with the provided context.
// This method initializes all configured components and starts the main service loop.
func (g *Gateway) Start(ctx context.Context) {
	log.Debug("start was called")

	if g.Certbot != nil {
		go g.Certbot.renewCertificates(ctx, defaultTickerRenewCertificates)
	}

	g.Controller.generateConfigs()
	g.Controller.start()
}

func main() {
	loglevel.UseEnvs()

	ctx := context.Background()

	config, err := loadGatewayConfig()
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	log.WithField("config", config).Info("configuration loaded")

	controller := &NginxController{
		gatewayConfig: config,
		rootDir:       defaultNginxRootDir,
		templatesDir:  defaultNginxTemplateDir,
	}

	gateway := NewGateway(config, controller)

	log.Info("gateway created")

	if envs.IsDevelopment() {
		log.Info("gateway running in development mode")

		log.Info("watch for nginx files is enabled")
		gateway.Watch()
	}

	if config.EnableAutoSSL {
		log.Info("auto ssl is enabled")

		gateway.EnableSSL()

		if config.Tunnels {
			log.WithFields(log.Fields{
				"provider": config.TunnelsDNSProvider,
				"token":    halfString(config.TunnelsDNSProviderToken),
			}).Info("tunnels info")

			gateway.EnableTunnels()
		}
	}

	log.Info("gateway started")
	gateway.Start(ctx)
}
