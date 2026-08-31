package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/shellhub-io/shellhub/agent/pkg/agentd"
	"github.com/shellhub-io/shellhub/agent/pkg/connector"
	"github.com/shellhub-io/shellhub/agent/pkg/selfupdater"
	"github.com/shellhub-io/shellhub/agent/server/modes/host/command"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/loglevel"
	"github.com/shellhub-io/shellhub/pkg/models"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{ //nolint:exhaustruct
		Use: "agent",
		Run: func(cmd *cobra.Command, _ []string) {
			loglevel.SetLogLevel()

			cfg, fields, err := agentd.LoadConfigFromEnv()
			if err != nil {
				log.WithError(err).WithFields(fields).Fatal("Failed to load de configuration from the environmental variables")
			}

			cfg.Version = AgentVersion
			cfg.Platform = AgentPlatform

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

			ag, err := agentd.NewAgentWithConfig(cfg, new(agentd.HostMode))
			if err != nil {
				log.WithError(err).WithFields(log.Fields{
					"version":       AgentVersion,
					"configuration": cfg,
				}).Fatal("Failed to create agent")
			}

			if err := ag.Setup(); err != nil {
				log.WithError(err).WithFields(log.Fields{
					"version":       AgentVersion,
					"configuration": cfg,
				}).Fatal("Failed to initialize agent")
			}

			if !cfg.HasNamespaceCredential() {
				tenant, err := waitForPairing(cmd.Context(), ag, cfg)
				if err != nil {
					log.WithError(err).Fatal("Failed to pair the device")
				}

				cfg.TenantID = tenant
				ag.SetTenantID(tenant)
			}

			if err := ag.Authorize(); err != nil {
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

	rootCmd.AddCommand(&cobra.Command{ //nolint:exhaustruct
		Use:   "info",
		Short: "Show information about the agent",
		Run: func(cmd *cobra.Command, _ []string) {
			loglevel.SetLogLevel()

			cfg, err := envs.ParseWithPrefix[agentd.Config]("SHELLHUB_")
			if err != nil {
				log.Fatal(err)
			}

			cfg.Version = AgentVersion
			cfg.Platform = AgentPlatform

			info, err := agentd.GetInfo(cfg)
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

			cmd.Println(string(data))
		},
	})

	rootCmd.AddCommand(&cobra.Command{ //nolint:exhaustruct
		Use:   "login",
		Short: "Accept this device from your browser",
		Long: `Prints a URL that accepts this device into its namespace. Open the URL in any
browser, sign in to the console, review the device and click Accept. The command
waits until the device is accepted, rejected, or the code expires.`,
		Run: func(cmd *cobra.Command, _ []string) {
			loglevel.SetLogLevel()

			cfg, fields, err := agentd.LoadConfigFromEnv()
			if err != nil {
				log.WithError(err).WithFields(fields).Fatal("Failed to load the configuration from the environmental variables")
			}

			cfg.Version = AgentVersion
			cfg.Platform = AgentPlatform

			if cfg.TenantID == "" {
				pairingLogin(cmd, cfg)

				return
			}

			ag, err := agentd.NewAgentWithConfig(cfg, new(agentd.HostMode))
			if err != nil {
				log.WithError(err).Fatal("Failed to create agent")
			}

			if err := ag.Initialize(); err != nil {
				log.WithError(err).Fatal("Failed to authenticate with the ShellHub server")
			}

			status, err := ag.DeviceStatus()
			if err != nil {
				log.WithError(err).Fatal("Failed to get the device status")
			}

			if status == models.DeviceStatusAccepted {
				cmd.Printf("✓ Device is already accepted into namespace %q.\n", ag.Namespace())

				return
			}

			code, err := ag.CreateDeviceLoginCode()
			if err != nil {
				log.WithError(err).Fatal("Failed to create the device login code")
			}

			printPairingInstructions(cmd, cfg.ServerAddress, code.Code, code.ExpiresIn)

			deadline := clock.Now().Add(time.Duration(code.ExpiresIn) * time.Second)
			for clock.Now().Before(deadline) {
				time.Sleep(3 * time.Second)

				status, err := ag.DeviceStatus()
				if err != nil {
					log.WithError(err).Warn("Failed to get the device status")

					continue
				}

				switch status {
				case models.DeviceStatusAccepted:
					cmd.Printf("✓ Device accepted into namespace %q.\n", ag.Namespace())

					return
				case models.DeviceStatusRejected:
					cmd.PrintErrln("✗ Device was rejected.")
					os.Exit(1)
				default:
				}
			}

			cmd.PrintErrln("✗ The code expired before the device was accepted. Run 'shellhub-agent login' again.")
			os.Exit(1)
		},
	})

	registerInstallerCommands(rootCmd)

	rootCmd.AddCommand(&cobra.Command{ //nolint:exhaustruct
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

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func waitForPairing(parent context.Context, ag *agentd.Agent, cfg *agentd.Config) (string, error) {
	ctx, stop := signal.NotifyContext(parent, os.Interrupt, syscall.SIGTERM)
	defer stop()

	tenantFile := agentd.TenantFilePath(cfg.PrivateKey)

	pairing, err := ag.CreatePairing()
	if err != nil {
		if cfg.PairingCode != "" {
			log.WithError(err).Warn("The pre-authorized pairing code was rejected; falling back to manual pairing")
			ag.ClearPairingCode()

			pairing, err = ag.CreatePairing()
		}

		if err != nil {
			return "", err
		}
	}

	if pairing.Status == models.DeviceStatusAccepted && pairing.TenantID != "" {
		if err := agentd.PersistTenant(tenantFile, pairing.TenantID); err != nil {
			log.WithError(err).Warn("Failed to persist the tenant; the device will need to be paired again on restart")
		}

		return pairing.TenantID, nil
	}

	base := strings.TrimRight(cfg.ServerAddress, "/")
	url := fmt.Sprintf("%s/accept-device?code=%s", base, pairing.Code)

	log.Info("This device is not enrolled in any namespace yet.")
	log.Infof("To pair it, open %s/accept-device and enter code %s", base, formatPairingCode(pairing.Code))
	log.Infof("Or open the link directly: %s", url)
	log.Infof("Waiting for acceptance... (code expires in %d minutes)", pairing.ExpiresIn/60)

	deadline := clock.Now().Add(time.Duration(pairing.ExpiresIn) * time.Second)
	for clock.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(3 * time.Second):
		}

		if tenant, err := agentd.ReadPersistedTenant(tenantFile); err == nil && tenant != "" {
			return tenant, nil
		}

		status, err := ag.GetPairingStatus(pairing.Code)
		if err != nil {
			log.WithError(err).Warn("Failed to get the pairing status")

			continue
		}

		if status.Status == models.DeviceStatusAccepted && status.TenantID != "" {
			if err := agentd.PersistTenant(tenantFile, status.TenantID); err != nil {
				log.WithError(err).Warn("Failed to persist the tenant; the device will need to be paired again on restart")
			}

			return status.TenantID, nil
		}
	}

	log.Info("The pairing code expired. Run 'shellhub-agent login' to get a new one.")

	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(10 * time.Second):
		}

		if tenant, err := agentd.ReadPersistedTenant(tenantFile); err == nil && tenant != "" {
			return tenant, nil
		}
	}
}

func pairingLogin(cmd *cobra.Command, cfg *agentd.Config) {
	ag, err := agentd.NewAgentWithConfig(cfg, new(agentd.HostMode))
	if err != nil {
		log.WithError(err).Fatal("Failed to create agent")
	}

	if err := ag.Setup(); err != nil {
		log.WithError(err).Fatal("Failed to reach the ShellHub server")
	}

	pairing, err := ag.CreatePairing()
	if err != nil {
		log.WithError(err).Fatal("Failed to create the device pairing code")
	}

	tenantFile := agentd.TenantFilePath(cfg.PrivateKey)

	if pairing.Status == models.DeviceStatusAccepted && pairing.TenantID != "" {
		if err := agentd.PersistTenant(tenantFile, pairing.TenantID); err != nil {
			log.WithError(err).Fatal("Device already accepted, but failed to persist the tenant")
		}

		cmd.Println("✓ This device is already accepted. The agent will connect automatically.")

		return
	}

	printPairingInstructions(cmd, cfg.ServerAddress, pairing.Code, pairing.ExpiresIn)

	deadline := clock.Now().Add(time.Duration(pairing.ExpiresIn) * time.Second)
	for clock.Now().Before(deadline) {
		time.Sleep(3 * time.Second)

		if tenant, err := agentd.ReadPersistedTenant(tenantFile); err == nil && tenant != "" {
			cmd.Println("✓ Device accepted. The agent will connect automatically.")

			return
		}

		status, err := ag.GetPairingStatus(pairing.Code)
		if err != nil {
			log.WithError(err).Warn("Failed to get the pairing status")

			continue
		}

		if status.Status == models.DeviceStatusAccepted && status.TenantID != "" {
			if err := agentd.PersistTenant(tenantFile, status.TenantID); err != nil {
				log.WithError(err).Fatal("Device accepted, but failed to persist the tenant")
			}

			cmd.Println("✓ Device accepted. The agent will connect automatically.")

			return
		}
	}

	cmd.PrintErrln("✗ The code expired before the device was accepted. Run 'shellhub-agent login' again.")
	os.Exit(1)
}

func formatPairingCode(code string) string {
	if len(code) != 8 {
		return code
	}

	return code[:4] + "-" + code[4:]
}

func printPairingInstructions(cmd *cobra.Command, serverAddress, code string, expiresInSeconds int) {
	base := strings.TrimRight(serverAddress, "/")
	url := fmt.Sprintf("%s/accept-device?code=%s", base, code)

	opened := openBrowser(url)

	cmd.Println()
	cmd.Println("To accept this device, open the ShellHub console and enter this code:")
	cmd.Println()
	cmd.Printf("      %s\n", formatPairingCode(code))
	cmd.Println()

	if opened {
		cmd.Println("We opened your browser. If nothing happened, open:")
	} else {
		cmd.Printf("Enter it at %s/accept-device, or open the link:\n", base)
	}

	cmd.Println()
	cmd.Printf("  %s\n", url)
	cmd.Println()
	cmd.Printf("Waiting for acceptance... (code expires in %d minutes)\n", expiresInSeconds/60)
}

func openBrowser(url string) bool {
	xdgOpen, err := exec.LookPath("xdg-open")
	if err != nil {
		return false
	}

	cmd := exec.Command(xdgOpen, url) //nolint:noctx,gosec // #nosec G204 -- xdgOpen comes from LookPath and url is built from local config
	if err := cmd.Start(); err != nil {
		return false
	}

	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	select {
	case err := <-done:
		return err == nil
	case <-time.After(2 * time.Second):
		return true
	}
}
