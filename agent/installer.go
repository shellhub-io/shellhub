//go:build installer

package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"

	"github.com/spf13/cobra"
)

const (
	agentEnvFile     = "/etc/shellhub-agent.env"
	agentServiceFile = "/etc/systemd/system/shellhub-agent.service"
	agentServiceName = "shellhub-agent"
)

var agentServiceTemplate = `[Unit]
Description=ShellHub Agent
After=local-fs.target network-online.target time-sync.target
Wants=network-online.target
Requires=local-fs.target

[Service]
EnvironmentFile=/etc/shellhub-agent.env
ExecStart={{.BinaryPath}}
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
`

type installerConfig struct {
	ServerAddress     string
	TenantID          string
	PrivateKey        string
	PreferredHostname string
	PreferredIdentity string
	KeepaliveInterval uint
}

func registerInstallerCommands(rootCmd *cobra.Command) {
	installCmd := &cobra.Command{ // nolint: exhaustruct
		Use:   "install",
		Short: "Install ShellHub agent as a systemd service",
		RunE: func(cmd *cobra.Command, _ []string) error {
			flags := cmd.Flags()

			serverAddress, _ := flags.GetString("server-address")
			tenantID, _ := flags.GetString("tenant-id")
			privateKey, _ := flags.GetString("private-key")
			preferredHostname, _ := flags.GetString("preferred-hostname")
			preferredIdentity, _ := flags.GetString("preferred-identity")

			var keepaliveInterval uint
			if flags.Changed("keepalive-interval") {
				keepaliveInterval, _ = flags.GetUint("keepalive-interval")
			}

			if err := agentInstall(installerConfig{
				ServerAddress:     serverAddress,
				TenantID:          tenantID,
				PrivateKey:        privateKey,
				PreferredHostname: preferredHostname,
				PreferredIdentity: preferredIdentity,
				KeepaliveInterval: keepaliveInterval,
			}); err != nil {
				return fmt.Errorf("installation failed: %w", err)
			}

			fmt.Println("ShellHub agent installed successfully.")
			fmt.Println("Check status: systemctl status shellhub-agent")
			fmt.Println("View logs:    journalctl -u shellhub-agent -f")

			return nil
		},
	}

	installCmd.Flags().String("server-address", "", "ShellHub server address")
	installCmd.Flags().String("tenant-id", "", "Namespace tenant ID")
	installCmd.Flags().String("private-key", "/etc/shellhub.key", "Path to the agent private key file")
	installCmd.Flags().String("preferred-hostname", "", "Preferred device hostname")
	installCmd.Flags().String("preferred-identity", "", "Preferred device identity")
	installCmd.Flags().Uint("keepalive-interval", 30, "Keepalive interval in seconds")
	installCmd.MarkFlagRequired("server-address") //nolint:errcheck
	installCmd.MarkFlagRequired("tenant-id")       //nolint:errcheck

	rootCmd.AddCommand(installCmd)

	rootCmd.AddCommand(&cobra.Command{ // nolint: exhaustruct
		Use:   "uninstall",
		Short: "Uninstall ShellHub agent systemd service",
		RunE: func(_ *cobra.Command, _ []string) error {
			if err := agentUninstall(); err != nil {
				return fmt.Errorf("uninstallation failed: %w", err)
			}

			fmt.Println("ShellHub agent uninstalled successfully.")
			fmt.Println("Note: the binary and private key were not removed.")

			return nil
		},
	})
}

func agentInstall(cfg installerConfig) error {
	if os.Geteuid() != 0 {
		return fmt.Errorf("must be run as root")
	}

	if err := exec.Command("systemctl", "show-environment").Run(); err != nil {
		return fmt.Errorf("systemd is not available on this system")
	}

	// Stop existing service before overwriting files (re-install / upgrade).
	// Ignore error — service may not exist yet.
	exec.Command("systemctl", "disable", "--now", agentServiceName).Run() //nolint:errcheck

	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to resolve executable path: %w", err)
	}

	binaryPath, err := filepath.EvalSymlinks(exe)
	if err != nil {
		return fmt.Errorf("failed to resolve symlinks: %w", err)
	}

	if err := writeAgentEnvFile(cfg); err != nil {
		return fmt.Errorf("failed to write env file: %w", err)
	}

	if err := writeAgentServiceFile(binaryPath); err != nil {
		return fmt.Errorf("failed to write service file: %w", err)
	}

	if err := exec.Command("systemctl", "daemon-reload").Run(); err != nil {
		return fmt.Errorf("failed to reload systemd daemon: %w", err)
	}

	if err := exec.Command("systemctl", "enable", "--now", agentServiceName).Run(); err != nil {
		return fmt.Errorf("failed to enable service: %w", err)
	}

	return nil
}

func writeAgentEnvFile(cfg installerConfig) error {
	var buf bytes.Buffer

	fmt.Fprintf(&buf, "SHELLHUB_SERVER_ADDRESS=%s\n", cfg.ServerAddress)
	fmt.Fprintf(&buf, "SHELLHUB_TENANT_ID=%s\n", cfg.TenantID)
	fmt.Fprintf(&buf, "SHELLHUB_PRIVATE_KEY=%s\n", cfg.PrivateKey)

	if cfg.PreferredHostname != "" {
		fmt.Fprintf(&buf, "SHELLHUB_PREFERRED_HOSTNAME=%s\n", cfg.PreferredHostname)
	}

	if cfg.PreferredIdentity != "" {
		fmt.Fprintf(&buf, "SHELLHUB_PREFERRED_IDENTITY=%s\n", cfg.PreferredIdentity)
	}

	if cfg.KeepaliveInterval != 0 {
		fmt.Fprintf(&buf, "SHELLHUB_KEEPALIVE_INTERVAL=%d\n", cfg.KeepaliveInterval)
	}

	return os.WriteFile(agentEnvFile, buf.Bytes(), 0600)
}

func writeAgentServiceFile(binaryPath string) error {
	tmpl, err := template.New("service").Parse(agentServiceTemplate)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, struct{ BinaryPath string }{binaryPath}); err != nil {
		return err
	}

	return os.WriteFile(agentServiceFile, buf.Bytes(), 0644)
}

func agentUninstall() error {
	if os.Geteuid() != 0 {
		return fmt.Errorf("must be run as root")
	}

	// Ignore error — service may already be stopped or not exist.
	exec.Command("systemctl", "disable", "--now", agentServiceName).Run() //nolint:errcheck

	if err := os.Remove(agentServiceFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove service file: %w", err)
	}

	if err := os.Remove(agentEnvFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove env file: %w", err)
	}

	if err := exec.Command("systemctl", "daemon-reload").Run(); err != nil {
		return fmt.Errorf("failed to reload systemd daemon: %w", err)
	}

	return nil
}
