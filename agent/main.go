package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"runtime"
	"time"

	"github.com/Masterminds/semver"
	"github.com/shellhub-io/shellhub/agent/pkg/connector"
	"github.com/shellhub-io/shellhub/agent/pkg/selfupdater"
	"github.com/shellhub-io/shellhub/agent/server/modes/host/command"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/loglevel"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func main() {
	// Default command.
	rootCmd := &cobra.Command{ // nolint: exhaustruct
		Use: "agent",
		Run: func(cmd *cobra.Command, _ []string) {
			loglevel.SetLogLevel()

			cfg, fields, err := LoadConfigFromEnv()
			if err != nil {
				log.WithError(err).WithFields(fields).Fatal("Failed to load de configuration from the environmental variables")
			}

			if os.Geteuid() == 0 && cfg.SingleUserPassword != "" {
				log.Error("ShellHub agent cannot run as root when single-user mode is enabled.")
				log.Error("To disable single-user mode unset SHELLHUB_SINGLE_USER_PASSWORD env.")
				os.Exit(1)
			}

			if os.Geteuid() != 0 && cfg.SingleUserPassword == "" {
				log.Error("When running as non-root user you need to set password for single-user mode by SHELLHUB_SINGLE_USER_PASSWORD environment variable.")
				log.Error("You can use openssl passwd utility to generate password hash. The following algorithms are supported: bsd1, apr1, sha256, sha512.")
				log.Error("Example: SHELLHUB_SINGLE_USER_PASSWORD=$(openssl passwd -6)")
				log.Error("See man openssl-passwd for more information.")
				os.Exit(1)
			}

			updater, err := selfupdater.NewUpdater(AgentVersion)
			if err != nil {
				log.Panic(err)
			}

			if err := updater.CompleteUpdate(); err != nil {
				log.Warning(err)
				os.Exit(0)
			}

			currentVersion := new(semver.Version)

			if AgentVersion != "latest" {
				currentVersion, err = updater.CurrentVersion()
				if err != nil {
					log.Panic(err)
				}
			}

			mode := func() string {
				if cfg.SingleUserPassword != "" {
					return "single-user"
				}

				return "multi-user"
			}()

			log.WithFields(log.Fields{
				"version": AgentVersion,
				"mode":    mode,
			}).Info("Starting ShellHub")

			ag, err := NewAgentWithConfig(cfg, new(HostMode))
			if err != nil {
				log.WithError(err).WithFields(log.Fields{
					"version":       AgentVersion,
					"configuration": cfg,
				}).Fatal("Failed to create agent")
			}

			if err := ag.Initialize(); err != nil {
				log.WithError(err).WithFields(log.Fields{
					"version":       AgentVersion,
					"configuration": cfg,
				}).Fatal("Failed to initialize agent")
			}

			ctx := cmd.Context()

			log.WithFields(log.Fields{
				"version":            AgentVersion,
				"mode":               mode,
				"tenant_id":          cfg.TenantID,
				"server_address":     cfg.ServerAddress,
				"preferred_hostname": cfg.PreferredHostname,
			}).Info("Listening for connections")

			// Disable check update in development mode
			if AgentVersion != "latest" {
				go func() {
					for {
						nextVersion, err := ag.CheckUpdate()
						if err != nil {
							log.WithError(err).WithFields(log.Fields{
								"version":            AgentVersion,
								"mode":               mode,
								"tenant_id":          cfg.TenantID,
								"server_address":     cfg.ServerAddress,
								"preferred_hostname": cfg.PreferredHostname,
							}).Error("Failed to check update")

							goto sleep
						}

						if nextVersion.GreaterThan(currentVersion) {
							if err := updater.ApplyUpdate(nextVersion); err != nil {
								log.WithError(err).WithFields(log.Fields{
									"version":            AgentVersion,
									"mode":               mode,
									"tenant_id":          cfg.TenantID,
									"server_address":     cfg.ServerAddress,
									"preferred_hostname": cfg.PreferredHostname,
								}).Error("Failed to apply update")
							}

							log.WithFields(log.Fields{
								"version":            currentVersion,
								"next_version":       nextVersion.String(),
								"mode":               mode,
								"tenant_id":          cfg.TenantID,
								"server_address":     cfg.ServerAddress,
								"preferred_hostname": cfg.PreferredHostname,
							}).Info("Update successfully applied")
						}

					sleep:
						log.WithFields(log.Fields{
							"version":            AgentVersion,
							"mode":               mode,
							"tenant_id":          cfg.TenantID,
							"server_address":     cfg.ServerAddress,
							"preferred_hostname": cfg.PreferredHostname,
						}).Info("Sleeping for 24 hours")

						time.Sleep(time.Hour * 24)
					}
				}()
			}

			if err := ag.Listen(ctx); err != nil {
				log.WithError(err).WithFields(log.Fields{
					"version":            AgentVersion,
					"mode":               mode,
					"tenant_id":          cfg.TenantID,
					"server_address":     cfg.ServerAddress,
					"preferred_hostname": cfg.PreferredHostname,
				}).Fatal("Failed to listen for connections")
			}

			log.WithFields(log.Fields{
				"version":            AgentVersion,
				"mode":               mode,
				"tenant_id":          cfg.TenantID,
				"server_address":     cfg.ServerAddress,
				"preferred_hostname": cfg.PreferredHostname,
			}).Info("Stopped listening for connections")
		},
	}

	rootCmd.AddCommand(&cobra.Command{
		Use:   "connector",
		Short: "Starts the ShellHub Agent in Connector mode",
		Run: func(cmd *cobra.Command, _ []string) {
			updater, err := selfupdater.NewUpdater(AgentVersion)
			if err != nil {
				log.Panic(err)
			}

			err = updater.CompleteUpdate()
			if err != nil {
				log.Warning(err)
				os.Exit(0)
			}

			currentVersion := new(semver.Version)

			if AgentVersion != "latest" {
				currentVersion, err = updater.CurrentVersion()
				if err != nil {
					log.Panic(err)
				}
			}

			cfg, fields, err := LoadConfigConnectorFromEnv()
			if err != nil {
				log.WithError(err).
					WithFields(fields).
					Fatal("Failed to load de configuration from the environmental variables")
			}

			logger := log.WithFields(
				log.Fields{
					"address":      cfg.ServerAddress,
					"tenant_id":    cfg.TenantID,
					"private_keys": cfg.PrivateKeys,
					"version":      AgentVersion,
				},
			)

			cfg.PrivateKeys = path.Dir(cfg.PrivateKeys)

			logger.Info("Starting ShellHub Agent Connector")

			connector.ConnectorVersion = AgentVersion
			connector, err := NewDockerConnector(cfg)
			if err != nil {
				logger.Fatal("Failed to create ShellHub Agent Connector")
			}

			if AgentVersion != "latest" {
				go func() {
					for {
						nextVersion, err := connector.CheckUpdate()
						if err != nil {
							log.WithError(err).WithFields(log.Fields{
								"version": AgentVersion,
							}).Error("Failed to check update")

							goto sleep
						}

						if nextVersion.GreaterThan(currentVersion) {
							if err := updater.ApplyUpdate(nextVersion); err != nil {
								log.WithError(err).WithFields(log.Fields{
									"version": AgentVersion,
								}).Error("Failed to apply update")
							}

							log.WithFields(log.Fields{
								"version":      currentVersion,
								"next_version": nextVersion.String(),
							}).Info("Update successfully applied")
						}

					sleep:
						log.WithFields(log.Fields{
							"version": AgentVersion,
						}).Info("Sleeping for 24 hours")

						time.Sleep(time.Hour * 24)
					}
				}()
			}

			if err := connector.Listen(cmd.Context()); err != nil {
				logger.Fatal("Failed to listen for connections")
			}

			logger.Info("ShellHub Agent Connector stopped")
		},
	})

	rootCmd.AddCommand(&cobra.Command{ // nolint: exhaustruct
		Use:   "info",
		Short: "Show information about the agent",
		Run: func(cmd *cobra.Command, _ []string) {
			loglevel.SetLogLevel()

			cfg, err := envs.ParseWithPrefix[Config]("SHELLHUB_")
			if err != nil {
				log.Fatal(err)
			}

			info, err := GetInfo(cfg)
			if err != nil {
				log.WithError(err).WithFields(log.Fields{
					"version":       AgentVersion,
					"configuration": cfg,
				}).Fatal("Failed to get agent information")
			}

			log.WithFields(log.Fields{
				"version": info.Version,
				"api":     info.Endpoints.API,
				"ssh":     info.Endpoints.SSH,
			}).Info("ShellHub agent information")

			data, err := json.Marshal(info)
			if err != nil {
				log.WithError(err).WithFields(log.Fields{
					"version":       AgentVersion,
					"configuration": cfg,
				}).Fatal("Failed to marshal agent information")
			}

			// NOTICE: this output was made to enable the agent's user to check and parse the agent's information with
			// a know format without having to parse the log output.
			// TODO: Should it have line break or not?
			cmd.Println(string(data))
		},
	})

	registerInstallerCommands(rootCmd)

	rootCmd.AddCommand(&cobra.Command{ // nolint: exhaustruct
		Use:   "sftp",
		Short: "Starts the SFTP server",
		Long: `Starts the SFTP server. This command is used internally by the agent and should not be used directly.
It is initialized by the agent when a new SFTP session is created.`,
		Run: func(_ *cobra.Command, args []string) {
			NewSFTPServer(command.SFTPServerMode(args[0]))
		},
	})

	rootCmd.Version = AgentVersion

	rootCmd.SetVersionTemplate(fmt.Sprintf("{{ .Name }} version: {{ .Version }}\ngo: %s\n",
		runtime.Version(),
	))

	rootCmd.Execute() // nolint: errcheck
}
