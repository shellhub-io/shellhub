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

// Run is the entry point for the API process. It parses environment variables,
// sets up the server, and blocks until a termination signal is received.
func Run() {
	loglevel.UseEnvs()

	// Both the server and the admin commands reach envs.IsCloud/IsEnterprise,
	// which panic on an unrecognized edition. Resolving here turns a typo into a
	// clean startup failure instead of a panic deep in a request.
	if _, err := envs.ResolveEdition(); err != nil {
		log.WithError(err).Fatal("failed to resolve ShellHub edition")
	}

	rootCmd := &cobra.Command{Use: "server"}
	rootCmd.AddCommand(admin.Command())
	rootCmd.AddCommand(&cobra.Command{
		Use: "server",
		RunE: func(cmd *cobra.Command, _ []string) error {
			env, err := envs.ParseWithPrefix[Env]("API_")
			if err != nil {
				log.WithError(err).
					Error("Failed to load environment variables")

				return err
			}

			server := &Server{env: env}

			if err := server.Setup(cmd.Context()); err != nil {
				log.WithError(err).
					Error("failed to setup the server")

				return err
			}

			sigs := make(chan os.Signal, 1)
			signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

			go func() {
				sig := <-sigs
				log.WithField("signal", sig).
					Info("shutting down the server")

				server.Shutdown()
				os.Exit(0)
			}()

			if err := server.Start(); err != nil {
				log.WithError(err).
					Error("failed too start the server")

				return err
			}

			return nil
		},
	})

	if err := rootCmd.ExecuteContext(context.Background()); err != nil {
		log.WithError(err).
			Error("failed to execute command")

		os.Exit(1)
	}
}
