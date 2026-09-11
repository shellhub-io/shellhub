package agentd

import (
	"crypto/rand"
	"crypto/rsa"
	"path/filepath"
	"testing"

	"github.com/pkg/errors"
	"github.com/shellhub-io/shellhub/agent/pkg/keygen"
	client_mocks "github.com/shellhub-io/shellhub/pkg/api/client/mocks"
	"github.com/shellhub-io/shellhub/pkg/envs"
	env_mocks "github.com/shellhub-io/shellhub/pkg/envs/mocks"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/validator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func ExampleNewAgentWithConfig() {
	_, err := NewAgentWithConfig(&Config{
		ServerAddress: "http://localhost:80",
		TenantID:      "00000000-0000-4000-0000-000000000000",
		PrivateKey:    "./shellhub.key",
	}, new(HostMode))
	if err != nil {
		panic(err)
	}
}

func ExampleNewAgent() {
	_, err := NewAgent("http://localhost:80", "00000000-0000-4000-0000-000000000000", "./shellhub.key", new(HostMode))
	if err != nil {
		panic(err)
	}
}

func TestLoadConfigFromEnv(t *testing.T) {
	envMock := new(env_mocks.MockBackend)
	envs.DefaultBackend = envMock

	type expected struct {
		cfg    *Config
		fields map[string]any
		err    error
	}

	tests := []struct {
		description   string
		requiredMocks func()
		expected      expected
	}{
		{
			description: "fail to load the environment variables when required ones are not set",
			requiredMocks: func() {
				envs := new(Config)

				envMock.On("Process", "SHELLHUB_", envs).Return(errors.New("")).Once()
			},
			expected: expected{
				cfg:    nil,
				fields: nil,
				err:    envs.ErrParseWithPrefix,
			},
		},
		{
			description: "fail to load the environment variables when one required values is empty",
			requiredMocks: func() {
				envs := new(Config)

				envMock.On("Process", "SHELLHUB_", envs).Return(nil).Once().Run(func(args mock.Arguments) {
					cfg, ok := args.Get(1).(*Config)
					require.True(t, ok)

					cfg.ServerAddress = "http://localhost"
					cfg.TenantID = ""
					cfg.PrivateKey = ""
					cfg.MaxRetryConnectionTimeout = 30
				})
			},
			expected: expected{
				cfg: nil,
				fields: map[string]any{
					"PrivateKey": "required",
				},
				err: validator.ErrStructureInvalid,
			},
		},
		{
			description: "fail to load the environment variables when required values are empty",
			requiredMocks: func() {
				envs := new(Config)

				envMock.On("Process", "SHELLHUB_", envs).Return(nil).Once().Run(func(args mock.Arguments) {
					cfg, ok := args.Get(1).(*Config)
					require.True(t, ok)

					cfg.ServerAddress = ""
					cfg.TenantID = ""
					cfg.PrivateKey = ""
					cfg.MaxRetryConnectionTimeout = 30
				})
			},
			expected: expected{
				cfg: nil,
				fields: map[string]any{
					"ServerAddress": "required",
					"PrivateKey":    "required",
				},
				err: validator.ErrStructureInvalid,
			},
		},
		{
			description: "fail to load the environment variables when the tenant is not a uuid",
			requiredMocks: func() {
				envs := new(Config)

				envMock.On("Process", "SHELLHUB_", envs).Return(nil).Once().Run(func(args mock.Arguments) {
					cfg, ok := args.Get(1).(*Config)
					require.True(t, ok)

					cfg.ServerAddress = "http://localhost"
					cfg.TenantID = "1c462afa-e4b6-41a5-ba54-7236a177O466"
					cfg.PrivateKey = "/tmp/shellhub.key"
					cfg.MaxRetryConnectionTimeout = 30
				})
			},
			expected: expected{
				cfg: nil,
				fields: map[string]any{
					"TenantID": "uuid",
				},
				err: validator.ErrStructureInvalid,
			},
		},
		{
			description: "fail to load the environment variables when the persisted tenant is not a uuid",
			requiredMocks: func() {
				key := filepath.Join(t.TempDir(), "shellhub.key")
				require.NoError(t, PersistTenant(TenantFilePath(key), "not-a-uuid"))

				envs := new(Config)

				envMock.On("Process", "SHELLHUB_", envs).Return(nil).Once().Run(func(args mock.Arguments) {
					cfg, ok := args.Get(1).(*Config)
					require.True(t, ok)

					cfg.ServerAddress = "http://localhost"
					cfg.TenantID = ""
					cfg.PrivateKey = key
					cfg.MaxRetryConnectionTimeout = 30
				})
			},
			expected: expected{
				cfg: nil,
				fields: map[string]any{
					"TenantID": "uuid",
				},
				err: validator.ErrStructureInvalid,
			},
		},
		{
			description: "success to load the environmental variables",
			requiredMocks: func() {
				envs := new(Config)

				envMock.On("Process", "SHELLHUB_", envs).Return(nil).Once().Run(func(args mock.Arguments) {
					cfg, ok := args.Get(1).(*Config)
					require.True(t, ok)

					cfg.ServerAddress = "http://localhost"
					cfg.TenantID = "1c462afa-e4b6-41a5-ba54-7236a1770466"
					cfg.PrivateKey = "/tmp/shellhub.key"
					cfg.MaxRetryConnectionTimeout = 30
				})
			},
			expected: expected{
				cfg: &Config{
					ServerAddress:             "http://localhost",
					TenantID:                  "1c462afa-e4b6-41a5-ba54-7236a1770466",
					TenantOrigin:              TenantFromEnvironment,
					PrivateKey:                "/tmp/shellhub.key",
					MaxRetryConnectionTimeout: 30,
				},
				fields: nil,
				err:    nil,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			test.requiredMocks()

			cfg, fields, err := LoadConfigFromEnv()
			assert.Equal(t, test.expected.cfg, cfg)
			assert.Equal(t, test.expected.fields, fields)
			assert.ErrorIs(t, err, test.expected.err)
		})
	}
}

func TestNewAgent(t *testing.T) {
	agent, err := NewAgent("http://localhost:80", "00000000-0000-4000-0000-000000000000", "./shellhub.key", new(HostMode))

	require.NoError(t, err)
	assert.Equal(t, TransportV2, agent.config.TransportVersion)
}

func TestNewAgentWithConfig(t *testing.T) {
	type expected struct {
		agent *Agent
		err   error
	}

	config := &Config{
		ServerAddress:    "http://localhost",
		TenantID:         "1c462afa-e4b6-41a5-ba54-7236a1770466",
		PrivateKey:       "/tmp/shellhub.key",
		TransportVersion: TransportV2,
	}

	tests := []struct {
		description string
		config      *Config
		mode        Mode
		expected    expected
	}{
		{
			description: "fail when server address is empty",
			config: &Config{
				ServerAddress: "",
			},
			mode: new(HostMode),
			expected: expected{
				agent: nil,
				err:   ErrNewAgentWithConfigEmptyServerAddress,
			},
		},
		{
			description: "fail when server address is invalid",
			config: &Config{
				ServerAddress: "invalid_url",
			},
			mode: new(HostMode),
			expected: expected{
				agent: nil,
				err:   ErrNewAgentWithConfigInvalidServerAddress,
			},
		},
		{
			description: "fail when private key is empty",
			config: &Config{
				ServerAddress: "http://localhost",
				TenantID:      "",
				PrivateKey:    "",
			},
			mode: new(HostMode),
			expected: expected{
				agent: nil,
				err:   ErrNewAgentWithConfigEmptyPrivateKey,
			},
		},
		{
			description: "fail when mode is nil",
			config: &Config{
				ServerAddress:    "http://localhost",
				TenantID:         "1c462afa-e4b6-41a5-ba54-7236a1770466",
				PrivateKey:       "/tmp/shellhub.key",
				TransportVersion: TransportV2,
			},
			mode: nil,
			expected: expected{
				agent: nil,
				err:   ErrNewAgentWithConfigNilMode,
			},
		},
		{
			description: "fail when transport version is unsupported",
			config: &Config{
				ServerAddress:    "http://localhost",
				TenantID:         "1c462afa-e4b6-41a5-ba54-7236a1770466",
				PrivateKey:       "/tmp/shellhub.key",
				TransportVersion: 0,
			},
			mode: new(HostMode),
			expected: expected{
				agent: nil,
				err:   ErrNewAgentWithConfigUnsupportedTransportVersion,
			},
		},
		{
			description: "success to create agent with transport version 1",
			config: &Config{
				ServerAddress:    "http://localhost",
				TenantID:         "1c462afa-e4b6-41a5-ba54-7236a1770466",
				PrivateKey:       "/tmp/shellhub.key",
				TransportVersion: TransportV1,
			},
			mode: new(HostMode),
			expected: expected{
				agent: &Agent{
					config: &Config{
						ServerAddress:    "http://localhost",
						TenantID:         "1c462afa-e4b6-41a5-ba54-7236a1770466",
						PrivateKey:       "/tmp/shellhub.key",
						TransportVersion: TransportV1,
					},
					mode: new(HostMode),
				},
				err: nil,
			},
		},
		{
			description: "success to create agent with config",
			config:      config,
			mode:        new(HostMode),
			expected: expected{
				agent: &Agent{
					config: config,
					mode:   new(HostMode),
				},
				err: nil,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			agent, err := NewAgentWithConfig(test.config, test.mode)

			assert.Equal(t, test.expected.agent, agent)
			assert.ErrorIs(t, err, test.expected.err)
		})
	}
}

func TestAgent_GetInfo(t *testing.T) {
	clientMocks := new(client_mocks.MockClient)

	type expected struct {
		info *models.Info
		err  error
	}

	agent := &Agent{
		cli:    clientMocks,
		config: &Config{Version: "latest"},
	}

	err := errors.New("")

	tests := []struct {
		description   string
		requiredMocks func()
		expected      expected
	}{
		{
			description: "fail to get the server info",
			requiredMocks: func() {
				clientMocks.On("GetInfo", "latest").Return(nil, err).Once()
			},
			expected: expected{
				info: nil,
				err:  err,
			},
		},
		{
			description: "success to get the server info",
			requiredMocks: func() {
				clientMocks.On("GetInfo", "latest").Return(&models.Info{
					Version: "latest",
				}, nil).Once()
			},
			expected: expected{
				info: &models.Info{
					Version: "latest",
				},
				err: nil,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			test.requiredMocks()

			info, err := agent.GetInfo()

			assert.Equal(t, test.expected.info, info)
			assert.ErrorIs(t, err, test.expected.err)
		})
	}
}

func TestAgent_probeServerInfo(t *testing.T) {
	type expected struct {
		serverInfo *models.Info
		err        error
	}

	failure := errors.New("the server could not be reached")

	tests := []struct {
		description   string
		requiredMocks func(cli *client_mocks.MockClient)
		expected      expected
	}{
		{
			description: "leaves the server info unset when the probe fails",
			requiredMocks: func(cli *client_mocks.MockClient) {
				cli.On("GetInfo", "latest").Return(nil, failure).Once()
			},
			expected: expected{
				serverInfo: nil,
				err:        failure,
			},
		},
		{
			description: "stores the server info when the probe succeeds",
			requiredMocks: func(cli *client_mocks.MockClient) {
				cli.On("GetInfo", "latest").Return(&models.Info{Version: "latest"}, nil).Once()
			},
			expected: expected{
				serverInfo: &models.Info{Version: "latest"},
				err:        nil,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			cli := new(client_mocks.MockClient)
			test.requiredMocks(cli)

			agent := &Agent{
				cli:    cli,
				config: &Config{Version: "latest"},
			}

			require.ErrorIs(t, agent.probeServerInfo(), test.expected.err)
			assert.Equal(t, test.expected.serverInfo, agent.serverInfo)
		})
	}
}

// TestAgent_generatePrivateKey_PathContainment verifies that the production
// generatePrivateKey method rejects PrivateKey paths that contain raw ".."
// traversal sequences.  The raw path is what an operator would supply via
// the PRIVATE_KEY environment variable, so filepath.Join is intentionally
// NOT used here — it would silently clean the traversal before the test runs.
func TestAgent_generatePrivateKey_PathContainment(t *testing.T) {
	t.Parallel()

	baseDir := t.TempDir()

	tests := []struct {
		description string
		privateKey  string
		wantErr     error
	}{
		{
			description: "valid absolute path with no traversal",
			privateKey:  baseDir + "/device.key",
			wantErr:     nil,
		},
		{
			description: "path traversal via raw .. sequence",
			privateKey:  baseDir + "/../escaped.key",
			wantErr:     keygen.ErrPathTraversal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			t.Parallel()

			a := &Agent{config: &Config{PrivateKey: tt.privateKey}}
			err := a.generatePrivateKey()

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestAgent_readPublicKey_PathContainment verifies that the production
// readPublicKey method rejects PrivateKey paths that contain raw ".."
// traversal sequences.
func TestAgent_readPublicKey_PathContainment(t *testing.T) {
	t.Parallel()

	baseDir := t.TempDir()
	keyPath := baseDir + "/device.key"
	require.NoError(t, keygen.GeneratePrivateKey(keyPath))

	tests := []struct {
		description string
		privateKey  string
		wantErr     error
	}{
		{
			description: "valid absolute path with no traversal",
			privateKey:  keyPath,
			wantErr:     nil,
		},
		{
			description: "path traversal via raw .. sequence",
			privateKey:  baseDir + "/../escaped.key",
			wantErr:     keygen.ErrPathTraversal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			t.Parallel()

			a := &Agent{config: &Config{PrivateKey: tt.privateKey}}
			err := a.readPublicKey()

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestConfigHasNamespaceCredential(t *testing.T) {
	tests := []struct {
		description string
		config      *Config
		expected    bool
	}{
		{
			description: "has none when neither a tenant nor an install key is set",
			config:      &Config{},
			expected:    false,
		},
		{
			description: "has one when the install key is the only credential",
			config:      &Config{InstallKey: "00000000-0000-4000-0000-000000000000"},
			expected:    true,
		},
		{
			description: "has one when the tenant is the only credential",
			config:      &Config{TenantID: "00000000-0000-4000-0000-000000000000"},
			expected:    true,
		},
		{
			description: "has one when both are set, the key governing acceptance rather than enrollment",
			config: &Config{
				TenantID:   "00000000-0000-4000-0000-000000000000",
				InstallKey: "11111111-1111-4111-1111-111111111111",
			},
			expected: true,
		},
		{
			description: "has none when only a pairing code is set, since the code is claimed by the pairing flow",
			config:      &Config{PairingCode: "XDGBESC4"},
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.config.HasNamespaceCredential())
		})
	}
}

func TestAgentAuthorizeRequiresANamespaceCredential(t *testing.T) {
	agent := &Agent{config: &Config{}}

	assert.Equal(t, ErrAuthorizeNoNamespaceCredential, agent.Authorize())
}

func TestLoadConfigFromEnvRecordsTenantOrigin(t *testing.T) {
	const tenant = "1c462afa-e4b6-41a5-ba54-7236a1770466"

	tests := []struct {
		description string
		envTenant   string
		fileTenant  string
		expected    TenantOrigin
	}{
		{
			description: "no tenant anywhere leaves the origin unset",
			expected:    TenantFromNowhere,
		},
		{
			description: "a tenant from the environment is attributed to it",
			envTenant:   tenant,
			expected:    TenantFromEnvironment,
		},
		{
			description: "a tenant read from the file is attributed to the file",
			fileTenant:  tenant,
			expected:    TenantFromFile,
		},
		{
			description: "the environment wins, and keeps its own attribution",
			envTenant:   tenant,
			fileTenant:  "00000000-0000-4000-0000-000000000000",
			expected:    TenantFromEnvironment,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			envMock := new(env_mocks.MockBackend)
			envs.DefaultBackend = envMock

			key := filepath.Join(t.TempDir(), "shellhub.key")
			if test.fileTenant != "" {
				require.NoError(t, PersistTenant(TenantFilePath(key), test.fileTenant))
			}

			envMock.On("Process", "SHELLHUB_", new(Config)).Return(nil).Once().Run(func(args mock.Arguments) {
				cfg, ok := args.Get(1).(*Config)
				require.True(t, ok)

				cfg.ServerAddress = "http://localhost"
				cfg.TenantID = test.envTenant
				cfg.PrivateKey = key
				cfg.MaxRetryConnectionTimeout = 30
			})

			cfg, _, err := LoadConfigFromEnv()
			require.NoError(t, err)
			assert.Equal(t, test.expected, cfg.TenantOrigin)
		})
	}
}

func TestAuthorizeNamesTheCredentialItWasRefusedFor(t *testing.T) {
	refused := errors.New("namespace not found")

	tests := []struct {
		description string
		config      *Config
		expected    string
	}{
		{
			description: "a tenant an operator supplied names the variable it came from",
			config: &Config{
				TenantID:     "1c462afa-e4b6-41a5-ba54-7236a1770466",
				TenantOrigin: TenantFromEnvironment,
			},
			expected: "SHELLHUB_TENANT_ID",
		},
		{
			description: "a tenant left by a previous pairing names the file holding it",
			config: &Config{
				TenantID:     "1c462afa-e4b6-41a5-ba54-7236a1770466",
				TenantOrigin: TenantFromFile,
				PrivateKey:   "/etc/shellhub.key",
			},
			expected: "/etc/shellhub.key.tenant",
		},
		{
			description: "an install key is named rather than the tenant it would have resolved",
			config: &Config{
				InstallKey: "a-key",
			},
			expected: "install key",
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			cli := new(client_mocks.MockClient)
			cli.On("AuthDevice", mock.Anything).Return(nil, refused).Once()

			key, err := rsa.GenerateKey(rand.Reader, 2048)
			require.NoError(t, err)

			agent := &Agent{
				cli:      cli,
				config:   test.config,
				pubKey:   &key.PublicKey,
				Info:     new(models.DeviceInfo),
				Identity: &models.DeviceIdentity{MAC: "83:18:77:25:78:0d"},
			}

			err = agent.Authorize()
			require.Error(t, err)
			assert.Contains(t, err.Error(), test.expected)
			assert.ErrorIs(t, err, refused)
		})
	}
}

func TestInvalidConfigMessagesNamesTheEnvironmentVariable(t *testing.T) {
	cases := []struct {
		description string
		fields      map[string]any
		expected    []string
	}{
		{
			description: "names the variable and spells out the rule",
			fields:      map[string]any{"TenantID": "uuid"},
			expected:    []string{"SHELLHUB_TENANT_ID must be a UUID"},
		},
		{
			description: "reports a missing required variable",
			fields:      map[string]any{"ServerAddress": "required"},
			expected:    []string{"SHELLHUB_SERVER_ADDRESS is required"},
		},
		{
			description: "reports a value outside its range",
			fields:      map[string]any{"MaxRetryConnectionTimeout": "max"},
			expected:    []string{"SHELLHUB_MAX_RETRY_CONNECTION_TIMEOUT is out of range"},
		},
		{
			description: "orders the messages so a run is reproducible",
			fields:      map[string]any{"TenantID": "uuid", "ServerAddress": "required"},
			expected: []string{
				"SHELLHUB_SERVER_ADDRESS is required",
				"SHELLHUB_TENANT_ID must be a UUID",
			},
		},
		{
			description: "falls back to the field name when it reads no environment variable",
			fields:      map[string]any{"Version": "required"},
			expected:    []string{"Version is required"},
		},
		{
			description: "falls back to the rule's name when it has no plain wording",
			fields:      map[string]any{"TenantID": "startswith"},
			expected:    []string{"SHELLHUB_TENANT_ID is invalid (startswith)"},
		},
		{
			description: "reports nothing when nothing failed",
			fields:      nil,
			expected:    nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.Equal(t, tc.expected, InvalidConfigMessages(Config{}, tc.fields))
		})
	}
}
