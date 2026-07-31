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

	// `gateway config` prints what this environment produces and exits, so the
	// configuration can be read without turning on debug logging or guessing
	// from the outside what the proxy decided.
	if len(os.Args) > 1 && os.Args[1] == "config" {
		os.Stdout.Write(rendered) //nolint:errcheck

		return
	}

	// Deliberately not the whole struct: it carries the DNS provider's API
	// token, and this line would put it in the logs of every deployment that
	// serves web endpoints.
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

	// The configuration is built from the environment and never changes while
	// the process runs, so there is nothing to watch and nothing to reload: a
	// different configuration means a different container.
	//
	// Signals are handled here rather than through caddy.TrapSignals, whose own
	// documentation advises against it for a program that already owns its
	// lifecycle.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop

	log.Info("shutting down")

	if err := caddy.Stop(); err != nil {
		log.WithError(err).Error("failed to stop cleanly")
	}
}
