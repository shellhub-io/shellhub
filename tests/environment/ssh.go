package environment

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
)

// SSHHostKey returns the public half of the host key the server's SSH listener presents, read
// from the key file mounted into the server container. ctx bounds the copy out of the container.
// It returns the error of a copy that fails, which it does when the container is stopped or the
// file is missing, the error of a read that fails, and an error when the file does not hold a
// private key ssh can parse.
func (s *Stack) SSHHostKey(ctx context.Context) (ssh.PublicKey, error) {
	reader, err := s.Service(ServiceServer).CopyFileFromContainer(ctx, "/run/secrets/ssh_private_key")
	if err != nil {
		return nil, err
	}

	defer reader.Close() //nolint:errcheck // the key is already read; close is best-effort

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	signer, err := ssh.ParsePrivateKey(data)
	if err != nil {
		return nil, err
	}

	return signer.PublicKey(), nil
}

// SSHHostKey returns the host key the server's SSH listener presents, failing t if it cannot be
// read. See [Stack.SSHHostKey].
func (dc *DockerCompose) SSHHostKey(t *testing.T) ssh.PublicKey {
	t.Helper()

	key, err := dc.stack.SSHHostKey(t.Context())
	require.NoError(t, err)

	return key
}

// AwaitDeviceOffline waits until the device uid reports itself offline, which happens once the
// server sees its last tunnel close, at the latest on the tunnel's next missed ping. It fails t
// when the device is still online after 90 seconds.
func (dc *DockerCompose) AwaitDeviceOffline(t *testing.T, uid string) {
	t.Helper()

	require.EventuallyWithT(t, func(tt *assert.CollectT) {
		current, resp, err := dc.GetDevice(t.Context(), uid)
		if !assert.NoError(tt, err) {
			return
		}

		assert.Equal(tt, 200, resp.StatusCode(), resp.String())
		assert.False(tt, current.Online)
	}, 90*time.Second, 1*time.Second)
}
