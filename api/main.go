package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/loglevel"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func main() {
	loglevel.UseEnvs()

	rootCmd := &cobra.Command{Use: "api"}
	rootCmd.AddCommand(&cobra.Command{
		Use: "server",
		RunE: func(cmd *cobra.Command, _ []string) error {
			server := new(Server)

			env, err := envs.ParseWithPrefix[env]("API_")
			if err != nil {
				log.WithError(err).
					Error("Failed to load environment variables")

				return err
			}

			if err := server.Setup(cmd.Context(), env); err != nil {
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
