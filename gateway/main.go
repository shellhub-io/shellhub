package main

import (
	"context"
	"slices"
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

const (
	// SSLFeature indicates that SSL's feature is eanbled
	SSLFeature = "ssl"
	// WebEndpointsFeature indicates that WebEndpoints' feature is eanbled.
	WebEndpointsFeature = "feature"
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

	// Features contains feature flags to gateway.
	Features []string
}

// NewGateway creates a new Gateway instance with the provided configuration and controller.
// The Certbot component is initially set to nil and will be initialized only when
// SSL is explicitly enabled through EnableSSL().
func NewGateway(config *GatewayConfig, controller *NginxController, features []string) *Gateway {
	g := &Gateway{
		Config:     config,
		Controller: controller,
		Certbot:    nil,
	}

	// NOTE: [SSLFeature] indicates that SSL's feature is eanbled, configuring SSL certificate management for the
	// gateway. It sets up Certbot with the gateway's domain configuration and establishes automatic certificate
	// provisioning and renewal. The renewal callback is configured to reload Nginx when certificates are
	// renewed, ensuring the server uses the latest certificates without manual intervention.
	if slices.Contains(features, SSLFeature) {
		g.Certbot = newCertBot(&Config{
			RootDir:         defaultCertBotRootDir,
			RenewedCallback: g.Controller.reload,
		})

		g.Certbot.Certificates = append(
			g.Certbot.Certificates,
			NewDefaultCertificate(g.Config.Domain),
		)
	}

	// NOTE: [WebEndpointsFeature] indicates that WebEndpoints' feature is enabled, configuring necessary values to work with
	// SSL's enabled.
	if slices.Contains(features, WebEndpointsFeature) {
		if g.Certbot != nil {
			if g.Config.WebEndpointsDomain == "" {
				g.Config.WebEndpointsDomain = g.Config.Domain
			}

			g.Certbot.Certificates = append(
				g.Certbot.Certificates,
				NewWebEndpointsCertificate(
					g.Config.WebEndpointsDomain,
					g.Config.WebEndpointsDNSProvider,
					g.Config.WebEndpointsDNSProviderToken,
				),
			)
		}
	}

	return g
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
		g.Certbot.ensureCertificates()
		g.Certbot.executeRenewCertificates()

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

	features := []string{}

	if config.EnableAutoSSL {
		log.WithFields(log.Fields{
			"provider": config.WebEndpointsDNSProvider,
			"token":    halfString(config.WebEndpointsDNSProviderToken),
		}).Info("auto ssl is enabled")

		features = append(features, SSLFeature)
	}

	if config.WebEndpoints {
		log.WithFields(log.Fields{
			"provider": config.WebEndpointsDNSProvider,
			"token":    halfString(config.WebEndpointsDNSProviderToken),
		}).Info("tunnels info")

		features = append(features, WebEndpointsFeature)
	}

	gateway := NewGateway(config, controller, features)

	log.Info("gateway created")

	if envs.IsDevelopment() {
		log.Info("gateway running in development mode")

		log.Info("watch for nginx files is enabled")
		gateway.Watch()
	}

	log.Info("gateway started")
	gateway.Start(ctx)
}
