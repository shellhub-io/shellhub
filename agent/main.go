package main

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/Masterminds/semver"
	"github.com/shellhub-io/shellhub/agent/ssh/modes/host/command"
	"github.com/shellhub-io/shellhub/agent/updater"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/loglevel"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// AgentVersion store the version to be embed inside the binary. This is
// injected using `-ldflags` build option.
//
//	go build -ldflags "-X main.AgentVersion=1.2.3"
//
// If set to `latest`, the auto-updating mechanism is disabled. This is intended
// to be used during development only.
var AgentVersion string

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

			updater, err := updater.NewUpdater(AgentVersion)
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

			ag, err := NewAgentWithConfig(cfg)
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

			// NOTE: Disable check update in development mode.
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
		Run:   func(cmd *cobra.Command, _ []string) {},
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
