package main

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/Masterminds/semver"
	"github.com/kelseyhightower/envconfig"
	"github.com/shellhub-io/shellhub/pkg/agent"
	"github.com/shellhub-io/shellhub/pkg/agent/pkg/selfupdater"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/loglevel"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// AgentVersion store the version to be embed inside the binary. This is
// injected using `-ldflags` build option (e.g: `go build -ldflags "-X
// main.AgentVersion=1.2.3"`).
//
// If set to `latest`, the auto-updating mechanism is disabled. This is intended
// to be used during development only.
var AgentVersion string

func main() {
	// Default command.
	rootCmd := &cobra.Command{ // nolint: exhaustruct
		Use: "agent",
		Run: func(cmd *cobra.Command, args []string) {
			loglevel.SetLogLevel()

			// NOTE(r): When T, the generic parameter, is a structure with required tag, the fallback for an
			// "unprefixed" parameter is used.
			//
			// For example,
			//
			// For the structure below, the parser will parse successfully when the variables exist with or without the
			// prefixes since the "required" tag is set to true.
			//
			//  SHELLHUB_TENANT_ID=00000000-0000-4000-0000-000000000000 SERVER_ADDRESS=http://127.0.0.1
			//  PRIVATE_KEY=/tmp/shellhub sudo -E ./agent
			//
			//  struct {
			//    ServerAddress string `env:"SERVER_ADDRESS,required"`
			//    PrivateKey string `env:"PRIVATE_KEY,required"`
			//    TenantID string `env:"TENANT_ID,required"`
			//  }
			//
			//  This behavior is driven by the [env] package. Check it out for more information.
			//
			// [env]: https://github.com/caarlos0/env
			cfg, err := envs.ParseWithPrefix[agent.Config]("SHELLHUB_")
			if err != nil {
				envconfig.Usage("shellhub", &cfg) // nolint:errcheck
				log.Fatal(err)
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

			ag, err := agent.NewAgentWithConfig(cfg)
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

			go func() {
				// NOTICE: We only start to ping the server when the agent is ready to accept connections.
				// It will make the agent ping to server after the ticker time set on ping function, what is 10 minutes
				// by default.

				if err := ag.Ping(ctx, 0); err != nil {
					log.WithError(err).WithFields(log.Fields{
						"version":        AgentVersion,
						"mode":           mode,
						"tenant_id":      cfg.TenantID,
						"server_address": cfg.ServerAddress,
					}).Fatal("Failed to ping server")
				}

				log.WithFields(log.Fields{
					"version":        AgentVersion,
					"mode":           mode,
					"tenant_id":      cfg.TenantID,
					"server_address": cfg.ServerAddress,
				}).Info("Stopped pinging server")
			}()

			log.WithFields(log.Fields{
				"version":        AgentVersion,
				"mode":           mode,
				"tenant_id":      cfg.TenantID,
				"server_address": cfg.ServerAddress,
			}).Info("Listening for connections")

			// Disable check update in development mode
			if AgentVersion != "latest" {
				go func() {
					for {
						nextVersion, err := ag.CheckUpdate()
						if err != nil {
							log.Error(err)

							goto sleep
						}

						if nextVersion.GreaterThan(currentVersion) {
							if err := updater.ApplyUpdate(nextVersion); err != nil {
								log.Error(err)
							}

							log.WithFields(log.Fields{
								"version":        currentVersion,
								"next_version":   nextVersion.String(),
								"mode":           mode,
								"tenant_id":      cfg.TenantID,
								"server_address": cfg.ServerAddress,
							}).Info("update successfully applied")
						}

					sleep:
						time.Sleep(time.Hour * 24)
					}
				}()
			}

			if err := ag.Listen(ctx); err != nil {
				log.WithError(err).WithFields(log.Fields{
					"version":        AgentVersion,
					"mode":           mode,
					"tenant_id":      cfg.TenantID,
					"server_address": cfg.ServerAddress,
				}).Fatal("Failed to listen for connections")
			}

			log.WithFields(log.Fields{
				"version":        AgentVersion,
				"mode":           mode,
				"tenant_id":      cfg.TenantID,
				"server_address": cfg.ServerAddress,
			}).Info("Stopped listening for connections")
		},
	}

	rootCmd.AddCommand(&cobra.Command{ // nolint: exhaustruct
		Use:   "info",
		Short: "Show information about the agent",
		Run: func(cmd *cobra.Command, args []string) {
			loglevel.SetLogLevel()

			cfg := agent.Config{}

			// Process unprefixed env vars for backward compatibility
			envconfig.Process("", &cfg) // nolint:errcheck

			if err := envconfig.Process("shellhub", &cfg); err != nil {
				// show envconfig usage help users to run agent
				envconfig.Usage("shellhub", &cfg) // nolint:errcheck
				log.Fatal(err)
			}

			ag, err := agent.NewAgentWithConfig(&cfg)
			if err != nil {
				log.WithError(err).WithFields(log.Fields{
					"version":       AgentVersion,
					"configuration": cfg,
				}).Fatal("Failed to create agent")
			}

			info, err := ag.GetInfo()
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
		Run: func(cmd *cobra.Command, args []string) {
			agent.NewSFTPServer()
		},
	})

	rootCmd.Version = AgentVersion

	rootCmd.SetVersionTemplate(fmt.Sprintf("{{ .Name }} version: {{ .Version }}\ngo: %s\n",
		runtime.Version(),
	))

	agent.AgentVersion = AgentVersion
	agent.AgentPlatform = AgentPlatform

	rootCmd.Execute() // nolint: errcheck
}
