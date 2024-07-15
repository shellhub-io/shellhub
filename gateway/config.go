package main

import (
	"context"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/sethvargo/go-envconfig"
	"runtime"
)

// GatewayConfig holds the configuration settings for the gateway.
type GatewayConfig struct {
	Env                  string `env:"SHELLHUB_ENV"`
	Domain               string `env:"SHELLHUB_DOMAIN,required" validate:"hostname"`
	PublicUrlDomain      string `env:"SHELLHUB_PUBLIC_URL_DOMAIN"`
	WorkerProcesses      string `env:"WORKER_PROCESSES,default=auto"`
	MaxWorkerOpenFiles   int    `env:"MAX_WORKER_OPEN_FILES,default=0"`
	MaxWorkerConnections int    `env:"MAX_WORKER_CONNECTIONS,default=16384"`
	BacklogSize          int    `env:"BACKLOG_SIZE"`
	EnableAutoSSL        bool   `env:"SHELLHUB_AUTO_SSL"`
	EnableProxyProtocol  bool   `env:"SHELLHUB_PROXY"`
	EnableEnterprise     bool   `env:"SHELLHUB_ENTERPRISE"`
	EnableCloud          bool   `env:"SHELLHUB_CLOUD"`
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

// ApplyDefaults sets default values for the GatewayConfig if not provided.
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
}
