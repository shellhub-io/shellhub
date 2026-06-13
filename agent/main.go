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

	"github.com/Masterminds/semver"
	"github.com/shellhub-io/shellhub/agent/pkg/agentd"
	"github.com/shellhub-io/shellhub/agent/pkg/connector"
	"github.com/shellhub-io/shellhub/agent/pkg/selfupdater"
	"github.com/shellhub-io/shellhub/agent/server/modes/host/command"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/loglevel"
	"github.com/shellhub-io/shellhub/pkg/models"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func main() {
	// Default command.
	rootCmd := &cobra.Command{ // nolint: exhaustruct
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

			if cfg.TenantID == "" {
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

			// NOTICE: this output was made to enable the agent's user to check and parse the agent's information with
			// a know format without having to parse the log output.
			// TODO: Should it have line break or not?
			cmd.Println(string(data))
		},
	})

	rootCmd.AddCommand(&cobra.Command{ // nolint: exhaustruct
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

			// Without a tenant (no env, nothing persisted) the device does not
			// exist yet; enroll it through a pairing where the user picks the
			// namespace on the accept page.
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

			url := fmt.Sprintf("%s/accept-device?code=%s", strings.TrimRight(cfg.ServerAddress, "/"), code.Code)

			// NOTE: cobra's cmd.Print* writes to stderr; the install.sh wrapper merges
			// stderr into its pipe (2>&1) to scan this output for the URL.
			//
			// Best effort: the agent usually runs headless (container, SSH), but when
			// a desktop is around this saves a copy-paste. The link is always printed.
			if openBrowser(url) {
				cmd.Println("Opened your browser. If nothing happened, open:")
			} else {
				cmd.Println("To accept this device, open:")
			}
			cmd.Println()
			cmd.Printf("  %s\n", url)
			cmd.Println()
			cmd.Printf("Waiting for acceptance... (code expires in %d minutes)\n", code.ExpiresIn/60)

			deadline := time.Now().Add(time.Duration(code.ExpiresIn) * time.Second)
			for time.Now().Before(deadline) {
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
				}
			}

			cmd.PrintErrln("✗ The code expired before the device was accepted. Run 'shellhub-agent login' again.")
			os.Exit(1)
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

// waitForPairing enrolls a tenant-less agent from the main process: it creates
// ONE pairing code, logs the accept URL and polls until a user accepts. When
// the code expires it does NOT create another one — a fleet of orphan agents
// must not hammer the server forever — and from then on it only watches the
// tenant file, which a concurrent `shellhub-agent login` writes on success.
func waitForPairing(parent context.Context, ag *agentd.Agent, cfg *agentd.Config) (string, error) {
	// Scope SIGTERM handling to the pairing wait so the loops below can unwind
	// cleanly (the root command runs with a non-cancellable context).
	ctx, stop := signal.NotifyContext(parent, os.Interrupt, syscall.SIGTERM)
	defer stop()

	tenantFile := agentd.TenantFilePath(cfg.PrivateKey)

	pairing, err := ag.CreatePairing()
	if err != nil {
		return "", err
	}

	// Resume: the server already has this device accepted (e.g. it crashed
	// after acceptance but before persisting the tenant). Learn the tenant now
	// without a new pairing.
	if pairing.Status == models.DeviceStatusAccepted && pairing.TenantID != "" {
		if err := agentd.PersistTenant(tenantFile, pairing.TenantID); err != nil {
			log.WithError(err).Warn("Failed to persist the tenant; the device will need to be paired again on restart")
		}

		return pairing.TenantID, nil
	}

	url := fmt.Sprintf("%s/accept-device?code=%s", strings.TrimRight(cfg.ServerAddress, "/"), pairing.Code)

	log.Info("This device is not enrolled in any namespace yet.")
	log.Infof("To accept it, open: %s", url)
	log.Infof("Waiting for acceptance... (code expires in %d minutes)", pairing.ExpiresIn/60)

	deadline := time.Now().Add(time.Duration(pairing.ExpiresIn) * time.Second)
	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(3 * time.Second):
		}

		// A concurrent `shellhub-agent login` may have been accepted with its
		// own code; the tenant file reconciles both.
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

	// Passive phase: no more server traffic, only the local tenant file.
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

// pairingLogin enrolls a tenant-less agent interactively: prints the accept
// URL, tries the browser, and waits until a user accepts the device into a
// namespace of their choice.
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

	// Resume: the device is already accepted server-side; persist the tenant
	// and exit so the agent connects on its next start.
	if pairing.Status == models.DeviceStatusAccepted && pairing.TenantID != "" {
		if err := agentd.PersistTenant(tenantFile, pairing.TenantID); err != nil {
			log.WithError(err).Fatal("Device already accepted, but failed to persist the tenant")
		}

		cmd.Println("✓ This device is already accepted. The agent will connect automatically.")

		return
	}

	url := fmt.Sprintf("%s/accept-device?code=%s", strings.TrimRight(cfg.ServerAddress, "/"), pairing.Code)

	if openBrowser(url) {
		cmd.Println("Opened your browser. If nothing happened, open:")
	} else {
		cmd.Println("To accept this device, open:")
	}
	cmd.Println()
	cmd.Printf("  %s\n", url)
	cmd.Println()
	cmd.Printf("Waiting for acceptance... (code expires in %d minutes)\n", pairing.ExpiresIn/60)

	deadline := time.Now().Add(time.Duration(pairing.ExpiresIn) * time.Second)
	for time.Now().Before(deadline) {
		time.Sleep(3 * time.Second)

		// The main process may have logged its own pairing URL at boot; if that
		// one was accepted instead, the tenant file reconciles both.
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

// openBrowser makes a best-effort attempt to open the URL in the user's
// browser via xdg-open. It reports whether the browser was (probably)
// opened; callers should always print the URL as a fallback.
func openBrowser(url string) bool {
	xdgOpen, err := exec.LookPath("xdg-open")
	if err != nil {
		return false
	}

	cmd := exec.Command(xdgOpen, url) // #nosec G204 -- xdgOpen comes from LookPath and url is built from local config
	if err := cmd.Start(); err != nil {
		return false
	}

	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	// xdg-open normally exits right away after handing the URL to the browser;
	// a non-zero exit (no display, no handler) means nothing was opened. If it
	// is still running after the timeout, it likely became the browser process
	// itself, so leave it alone and assume it worked.
	select {
	case err := <-done:
		return err == nil
	case <-time.After(2 * time.Second):
		return true
	}
}
