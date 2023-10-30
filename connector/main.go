package main

import (
	"path"

	"github.com/kelseyhightower/envconfig"
	"github.com/shellhub-io/shellhub/pkg/agent"
	"github.com/shellhub-io/shellhub/pkg/agent/connector"
	"github.com/shellhub-io/shellhub/pkg/envs"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// ConnectorVersion store the version to be embed inside the binary. This is
// injected using `-ldflags` build option (e.g: `go build -ldflags "-X
// main.ConnectorVersion=1.2.3"`).
var ConnectorVersion string

func main() {
	rootCmd := &cobra.Command{ // nolint: exhaustruct
		Use:   "docker",
		Short: "Starts the Docker Connector",
		Long:  "Starts the Docker Connector, a service what turns all containers in a docker engine into a ShelHub device",
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := envs.ParseWithPrefix[agent.ConfigConnector]("SHELLHUB_")
			if err != nil {
				envconfig.Usage("shellhub", &cfg) // nolint:errcheck
				log.Fatal(err)
			}

			cfg.PrivateKeys = path.Dir(cfg.PrivateKeys)

			log.WithFields(log.Fields{
				"address":      cfg.ServerAddress,
				"tenant_id":    cfg.TenantID,
				"private_keys": cfg.PrivateKeys,
				"version":      ConnectorVersion,
			}).Info("Starting ShellHub Docker Connector")

			connector.ConnectorVersion = ConnectorVersion
			connector, err := connector.NewDockerConnector(cfg.ServerAddress, cfg.TenantID, cfg.PrivateKeys)
			if err != nil {
				log.WithError(err).WithFields(log.Fields{
					"address":   cfg.ServerAddress,
					"tenant_id": cfg.TenantID,
					"version":   ConnectorVersion,
				}).Fatal("Failed to create ShellHub Docker Connector")
			}

			if err := connector.Listen(cmd.Context()); err != nil {
				log.WithError(err).WithFields(log.Fields{
					"address":   cfg.ServerAddress,
					"tenant_id": cfg.TenantID,
					"version":   ConnectorVersion,
				}).Fatal("Failed to listen for connections")
			}

			log.WithFields(log.Fields{
				"address":   cfg.ServerAddress,
				"tenant_id": cfg.TenantID,
				"version":   ConnectorVersion,
			}).Info("ShellHub Docker Connector stopped")
		},
	}

	rootCmd.Version = ConnectorVersion
	rootCmd.Execute() // nolint: errcheck
}
