package main

import (
	"net/http"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
)

// TestRenameNamespace renames the namespace an online device belongs to through the API. The name
// is half of every SSHID that reaches the device, so the rename moves the device's address with it.
func TestRenameNamespace(t *testing.T) {
	t.Parallel()

	compose := newSSHEnvironment(t, t.Context(), models.SSHAccessModeLegacy)
	_, device := startAcceptedAgent(t, t.Context(), compose)
	signer := registerDeviceKey(t, t.Context(), compose)

	const takenName = "takenspace"
	compose.NewNamespace(t, ShellHubUsername, takenName, uuid.Generate(), "")

	sshTo := func(namespace string) *ssh.ClientConfig {
		return &ssh.ClientConfig{ //nolint:exhaustruct // the remaining fields keep their defaults
			User:            ShellHubAgentUsername + "@" + namespace + "." + device.Name,
			Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec // the test stack's host key is ephemeral
		}
	}

	require.NoError(t, dialClient(t, t.Context(), compose.SSHAddress(), sshTo(ShellHubNamespaceName)).Close())

	rename := func(t *testing.T, name string) (*models.Namespace, *resty.Response) {
		t.Helper()

		renamed := new(models.Namespace)
		resp, err := compose.R(t.Context()).
			SetBody(map[string]string{"name": name}).
			SetResult(renamed).
			Put("/api/namespaces/" + ShellHubNamespace)
		require.NoError(t, err)

		return renamed, resp
	}

	t.Run("a name another namespace holds answers conflict, whatever its case", func(t *testing.T) {
		_, resp := rename(t, "TakenSpace")
		require.Equal(t, http.StatusConflict, resp.StatusCode(), resp.String())

		current := new(models.Namespace)
		got, err := compose.R(t.Context()).SetResult(current).Get("/api/namespaces/" + ShellHubNamespace)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, got.StatusCode(), got.String())
		assert.Equal(t, ShellHubNamespaceName, current.Name)
	})

	t.Run("the new name is stored lowercased", func(t *testing.T) {
		renamed, resp := rename(t, "RenamedSpace")
		require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())
		assert.Equal(t, "renamedspace", renamed.Name)
	})

	t.Run("the device answers at the new name and no longer at the old one", func(t *testing.T) {
		conn := dialClient(t, t.Context(), compose.SSHAddress(), sshTo("renamedspace"))
		require.NoError(t, conn.Close())

		mark := compose.ServerLogMark(t)
		oldAddress := sshTo(ShellHubNamespaceName)

		require.Error(t, handshakeWith(t.Context(), compose.SSHAddress(), oldAddress))
		compose.AwaitServerLogLine(t, mark, "failed to create the session", oldAddress.User)
	})
}
