package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig"
	"github.com/shellhub-io/shellhub/pkg/loglevel"
	log "github.com/sirupsen/logrus"
)

func main() {
	loglevel.UseEnvs()

	config, err := loadGatewayConfig()
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	rendered, err := Caddyfile(config)
	if err != nil {
		log.Fatalf("failed to build the configuration: %v", err)
	}

	if len(os.Args) > 1 && os.Args[1] == "config" {
		os.Stdout.Write(rendered) //nolint:errcheck

		return
	}

	log.WithFields(log.Fields{
		"domain":   config.Domain,
		"env":      config.Env,
		"auto_ssl": config.EnableAutoSSL,
	}).Info("configuration loaded")

	adapted, warnings, err := caddyconfig.GetAdapter("caddyfile").Adapt(rendered, nil)
	if err != nil {
		log.Fatalf("failed to adapt the configuration: %v", err)
	}

	for _, warning := range warnings {
		log.WithField("line", warning.Line).Warn(warning.Message)
	}

	if err := caddy.Load(adapted, true); err != nil {
		log.Fatalf("failed to start: %v", err)
	}

	log.Info("gateway started")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop

	log.Info("shutting down")

	if err := caddy.Stop(); err != nil {
		log.WithError(err).Error("failed to stop cleanly")
	}
}
