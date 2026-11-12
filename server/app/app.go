// Package app provides the ShellHub server as an importable library.
//
// Community Edition builds use this package directly via server/main.go.
// Enterprise builds use a separate main in the cloud repo that blank-imports
// cloud extension packages before calling [Run], so their init() functions
// register routes, workers, billing providers, and store wrappers.
package app

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/loglevel"
	"github.com/shellhub-io/shellhub/server/admin"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// Run is the entry point for the API process. It parses environment variables, sets up the
// server, and serves until SIGINT or SIGTERM. It then shuts the server down and returns once the
// shutdown has finished, so the process exits with status 0. A second signal during the shutdown
// ends the process at once. When the environment does not parse, or the server fails to set up,
// to start or to serve, Run logs the error and exits the process with status 1 without shutting
// the server down.
func Run() {
	loglevel.UseEnvs(envs.OS())

	rootCmd := &cobra.Command{Use: "server"}
	rootCmd.AddCommand(admin.Command())
	rootCmd.AddCommand(&cobra.Command{
		Use: "server",
		RunE: func(cmd *cobra.Command, _ []string) error {
			env, err := envs.ParseWithPrefix[Env](envs.OS(), "API_")
			if err != nil {
				log.WithError(err).
					Error("Failed to load environment variables")

				return err
			}

			ssh, err := envs.ParseWithPrefix[sshEnv](envs.OS(), "SSH_")
			if err != nil {
				log.WithError(err).
					Error("Failed to load the SSH environment variables")

				return err
			}

			server := &Server{env: env, sshEnv: ssh}

			if err := server.Setup(cmd.Context()); err != nil {
				log.WithError(err).
					Error("failed to setup the server")

				return err
			}

			ctx, stop := signal.NotifyContext(cmd.Context(), syscall.SIGINT, syscall.SIGTERM)
			defer stop()

			if err := server.Start(ctx); err != nil {
				log.WithError(err).
					Error("failed to start the server")

				return err
			}

			stop()

			log.WithField("cause", context.Cause(ctx)).
				Info("shutting down the server")

			server.Shutdown()

			return nil
		},
	})

	if err := rootCmd.ExecuteContext(context.Background()); err != nil {
		log.WithError(err).
			Error("failed to execute command")

		os.Exit(1)
	}
}
