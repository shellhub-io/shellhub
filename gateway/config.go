package main

import (
	"context"
	"fmt"
	"runtime"

	"github.com/go-playground/validator/v10"
	"github.com/sethvargo/go-envconfig"
)

// GatewayConfig holds the configuration settings for the gateway.
type GatewayConfig struct {
	Env                          string      `env:"SHELLHUB_ENV"`
	Domain                       string      `env:"SHELLHUB_DOMAIN,required" validate:"hostname"`
	WebEndpoints                 bool        `env:"SHELLHUB_WEB_ENDPOINTS,default=false"`
	WebEndpointsDomain           string      `env:"SHELLHUB_WEB_ENDPOINTS_DOMAIN"`
	WebEndpointsDNSProvider      DNSProvider `env:"SHELLHUB_WEB_ENDPOINTS_DNS_PROVIDER,default=digitalocean"`
	WebEndpointsDNSProviderToken string      `env:"SHELLHUB_WEB_ENDPOINTS_DNS_PROVIDER_TOKEN"`
	WebEndpointsAcmeDNSURL       string      `env:"SHELLHUB_WEB_ENDPOINTS_ACME_DNS_URL"`
	WebEndpointsAcmeDNSUsername  string      `env:"SHELLHUB_WEB_ENDPOINTS_ACME_DNS_USERNAME"`
	WebEndpointsAcmeDNSPassword  string      `env:"SHELLHUB_WEB_ENDPOINTS_ACME_DNS_PASSWORD"`
	WebEndpointsAcmeDNSSubdomain string      `env:"SHELLHUB_WEB_ENDPOINTS_ACME_DNS_SUBDOMAIN"`
	WorkerProcesses              string      `env:"WORKER_PROCESSES,default=auto"`
	MaxWorkerOpenFiles           int         `env:"MAX_WORKER_OPEN_FILES,default=0"`
	MaxWorkerConnections         int         `env:"MAX_WORKER_CONNECTIONS,default=16384"`
	BacklogSize                  int         `env:"BACKLOG_SIZE"`
	EnableAutoSSL                bool        `env:"SHELLHUB_AUTO_SSL"`
	EnableProxyProtocol          bool        `env:"SHELLHUB_PROXY"`
	EnableEnterprise             bool        `env:"SHELLHUB_ENTERPRISE"`
	EnableCloud                  bool        `env:"SHELLHUB_CLOUD"`
	Database                     string      `env:"SHELLHUB_DATABASE,default=mongo"`
	EnableAccessLogs             bool        `env:"SHELLHUB_GATEWAY_ACCESS_LOGS" default:"true"`
	APIRateLimit                 string      `env:"SHELLHUB_API_RATE_LIMIT,default=1000r/s"`
	APIRateLimitZoneSize         string      `env:"SHELLHUB_API_RATE_LIMIT_ZONE_SIZE,default=10m"`
	APIBurstSize                 string      `env:"SHELLHUB_API_BURST_SIZE,default=1"`
	APIBurstDelay                string      `env:"SHELLHUB_API_BURST_DELAY,default=nodelay"`
	// APIBackend is the backend service to use for API requests (api:8080 or cloud:8080)
	// Set dynamically based on EnableCloud/EnableEnterprise
	APIBackend string
}

var validate = validator.New()

// loadGatewayConfig loads and validates the configuration from environment variables.
func loadGatewayConfig() (*GatewayConfig, error) {
	var config GatewayConfig
	if err := envconfig.Process(context.Background(), &config); err != nil {
		return nil, err
	}

	config.applyDefaults()

	if err := validate.Struct(config); err != nil {
		return nil, err
	}

	return &config, nil
}

// applyDefaults sets default values for the GatewayConfig if not provided.
func (gc *GatewayConfig) applyDefaults() {
	if gc.WorkerProcesses == "auto" {
		gc.WorkerProcesses = fmt.Sprintf("%d", runtime.NumCPU())
	}

	if gc.MaxWorkerOpenFiles == 0 {
		gc.MaxWorkerOpenFiles = rlimitMaxNumFiles() - 1024
		if gc.MaxWorkerOpenFiles < 1024 {
			gc.MaxWorkerOpenFiles = 1024
		}
	}

	if gc.MaxWorkerConnections == 0 {
		gc.MaxWorkerConnections = int(float64(gc.MaxWorkerOpenFiles * 3.0 / 4))
	}

	gc.BacklogSize = getSysctl("net.core.somaxconn")

	// Cloud and enterprise features are now unified into the api binary.
	// All traffic always routes to api:8080 regardless of edition.
	gc.APIBackend = "api:8080"
}
