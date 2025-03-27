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

	log.Info("Starting ShellHub server")

	apiServer := api.New()
	if err := apiServer.Setup(ctx); err != nil {
		log.WithError(err).Fatal("failed to setup API server")
	}

	router := apiServer.Router()

	sshServer, err := ssh.New(router)
	if err != nil {
		log.WithError(err).Fatal("failed to setup SSH server")
	}

	errCh := make(chan error, 2)

	go func() {
		log.Info("Starting API server")
		errCh <- apiServer.Start()
	}()

	go func() {
		log.Info("Starting SSH server")
		errCh <- sshServer.Start()
	}()

	go func() {
		sig := <-sigs
		log.WithField("signal", sig).Info("received shutdown signal")
		apiServer.Shutdown()
		sshServer.Shutdown()
		cancel()
	}()

	if err := <-errCh; err != nil {
		log.WithError(err).Fatal("server exited with error")
	}
}
