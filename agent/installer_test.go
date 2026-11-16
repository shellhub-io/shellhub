//go:build installer

package main

import (
	"testing"

	"github.com/shellhub-io/shellhub/agent/pkg/agentd"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestInstallRefusesAPlaintextServerAddress(t *testing.T) {
	t.Setenv(agentd.AllowInsecureServerEnv, "")

	root := &cobra.Command{Use: "shellhub-agent"}
	registerInstallerCommands(root)
	root.SetArgs([]string{
		"install",
		"--server-address", "http://shellhub.example.com",
		"--tenant-id", "00000000-0000-4000-0000-000000000000",
	})

	require.ErrorIs(t, root.Execute(), agentd.ErrPlaintextServerAddress)
}

func TestCheckInstallServerAddress(t *testing.T) {
	tests := []struct {
		description string
		address     string
		flag        bool
		environment string
		allowed     bool
		err         error
	}{
		{description: "refuses plain http", address: "http://shellhub.example.com", err: agentd.ErrPlaintextServerAddress},
		{description: "accepts plain http with the flag", address: "http://shellhub.example.com", flag: true, allowed: true},
		{description: "accepts plain http with the environment variable", address: "http://shellhub.example.com", environment: "true", allowed: true},
		{description: "accepts https without the opt-in", address: "https://shellhub.example.com"},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			t.Setenv(agentd.AllowInsecureServerEnv, test.environment)

			allowed, err := checkInstallServerAddress(test.address, test.flag)

			require.ErrorIs(t, err, test.err)
			require.Equal(t, test.allowed, allowed)
		})
	}
}

func TestAgentEnvFileCarriesTheInsecureServerOptIn(t *testing.T) {
	allowed := string(agentEnvFileContents(installerConfig{
		ServerAddress:       "http://shellhub.example.com",
		AllowInsecureServer: true,
		PrivateKey:          "/etc/shellhub.key",
	}))
	require.Contains(t, allowed, "SHELLHUB_ALLOW_INSECURE_SERVER=true\n")

	refused := string(agentEnvFileContents(installerConfig{
		ServerAddress: "https://shellhub.example.com",
		PrivateKey:    "/etc/shellhub.key",
	}))
	require.NotContains(t, refused, "SHELLHUB_ALLOW_INSECURE_SERVER")
}
