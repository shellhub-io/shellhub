package agentd

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/agent/pkg/keygen"
	"github.com/shellhub-io/shellhub/pkg/api/client"
	client_mocks "github.com/shellhub-io/shellhub/pkg/api/client/mocks"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type idleMode struct{}

func (idleMode) Serve(*Agent) {}

func (idleMode) GetInfo() (*Info, error) { return &Info{ID: "debian", Name: "Debian"}, nil }

func removalTestAgent(t *testing.T, cli *client_mocks.MockClient, tenant string) *Agent {
	t.Helper()

	keyPath := filepath.Join(t.TempDir(), "shellhub.key")
	require.NoError(t, keygen.GeneratePrivateKey(keyPath))

	pubKey, err := keygen.ReadPublicKey(keyPath)
	require.NoError(t, err)

	return &Agent{
		config:     &Config{TenantID: tenant, PrivateKey: keyPath, TransportVersion: TransportV2, MaxRetryConnectionTimeout: 60},
		mode:       idleMode{},
		cli:        cli,
		pubKey:     pubKey,
		Identity:   &models.DeviceIdentity{MAC: "aa:bb:cc:dd:ee:01"},
		Info:       &models.DeviceInfo{},
		serverInfo: &models.Info{Endpoints: models.Endpoints{SSH: "localhost:22", API: "localhost:80"}},
	}
}

func TestAuthorizeReportsARemovedDevice(t *testing.T) {
	cli := client_mocks.NewMockClient(t)
	cli.On("AuthDevice", mock.Anything).Return(nil, client.ErrUnauthorized).Once()

	ag := removalTestAgent(t, cli, "00000000-0000-4000-0000-000000000000")

	assert.ErrorIs(t, ag.Authorize(), ErrDeviceRemoved)
}

func TestListenStopsWhenTheDeviceIsRemoved(t *testing.T) {
	cli := client_mocks.NewMockClient(t)
	cli.On("AuthDevice", mock.Anything).Return(&models.DeviceAuthResponse{UID: "uid", Token: "token", Name: "dev", Namespace: "ns"}, nil).Once()

	ag := removalTestAgent(t, cli, "00000000-0000-4000-0000-000000000000")
	require.NoError(t, ag.Authorize())

	cli.On("AuthDevice", mock.Anything).Return(nil, client.ErrUnauthorized).Once()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	assert.ErrorIs(t, ag.Listen(ctx), ErrDeviceRemoved, "the tunnel authorizes before it dials and stops on a removal")
}

func TestUnpairForgetsAPairedTenant(t *testing.T) {
	ag := removalTestAgent(t, client_mocks.NewMockClient(t), "")
	tenantFile := TenantFilePath(ag.config.PrivateKey)

	require.NoError(t, PersistTenant(tenantFile, "00000000-0000-4000-0000-000000000000"))
	ag.SetTenantID("00000000-0000-4000-0000-000000000000")

	require.NoError(t, ag.Unpair())

	assert.False(t, ag.config.HasNamespaceCredential(), "the agent goes back to pairing")
	_, err := os.Stat(tenantFile)
	assert.True(t, os.IsNotExist(err), "the persisted tenant is gone")
	assert.Equal(t, TenantFromNowhere, ag.config.TenantOrigin, "a forgotten tenant has no origin left to report")
}

func TestUnpairForgetsATenantAdoptedFromTheFile(t *testing.T) {
	ag := removalTestAgent(t, client_mocks.NewMockClient(t), "00000000-0000-4000-0000-000000000000")
	ag.config.TenantOrigin = TenantFromFile
	tenantFile := TenantFilePath(ag.config.PrivateKey)
	require.NoError(t, PersistTenant(tenantFile, "00000000-0000-4000-0000-000000000000"))

	require.NoError(t, ag.Unpair())

	assert.False(t, ag.config.HasNamespaceCredential(), "a tenant a previous pairing persisted is the agent's to forget")
	_, err := os.Stat(tenantFile)
	assert.True(t, os.IsNotExist(err), "the persisted tenant is gone")
}

func TestUnpairRefusesATenantFromTheEnvironment(t *testing.T) {
	ag := removalTestAgent(t, client_mocks.NewMockClient(t), "00000000-0000-4000-0000-000000000000")

	require.ErrorIs(t, ag.Unpair(), ErrTenantFromEnvironment)
	assert.Equal(t, "00000000-0000-4000-0000-000000000000", ag.config.TenantID)
}

func TestTheTenantCanBeReadWhilePairingSetsIt(t *testing.T) {
	ag := removalTestAgent(t, client_mocks.NewMockClient(t), "")

	done := make(chan struct{})
	go func() {
		defer close(done)

		for range 100 {
			ag.CredentialFields()
			_, _ = ag.buildDeviceAuth()
		}
	}()

	for range 100 {
		ag.SetTenantID("00000000-0000-4000-0000-000000000000")
		require.NoError(t, ag.Unpair())
	}

	<-done
}
