package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/shellhub-io/shellhub/server/api"
	"github.com/shellhub-io/shellhub/server/ssh"

	"github.com/shellhub-io/shellhub/pkg/loglevel"
	log "github.com/sirupsen/logrus"
)

func main() {
	loglevel.UseEnvs()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	apiServer := api.New()
	if err := apiServer.Setup(ctx); err != nil {
		log.WithError(err).Fatal("failed to setup API server")
	}

	sshServer, err := ssh.New()
	if err != nil {
		log.WithError(err).Fatal("failed to setup SSH server")
	}

	errs := make(chan error, 2)

	go func() {
		log.Info("Starting API server")
		if err := apiServer.Start(); err != nil {
			errs <- err
		}
	}()

	go func() {
		log.Info("Starting SSH server")
		if err := sshServer.Start(); err != nil {
			errs <- err
		}
	}()

	go func() {
		sig := <-sigs
		log.WithField("signal", sig).Info("received shutdown signal")

		apiServer.Shutdown()
		sshServer.Shutdown()
		cancel()
	}()

	if err := <-errs; err != nil {
		log.WithError(err).Fatal("server exited with error")
	}
}
