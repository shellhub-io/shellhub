package agent

import (
	"testing"

	"github.com/pkg/errors"
	client_mocks "github.com/shellhub-io/shellhub/pkg/api/client/mocks"
	"github.com/shellhub-io/shellhub/pkg/envs"
	env_mocks "github.com/shellhub-io/shellhub/pkg/envs/mocks"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/validator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func ExampleNewAgentWithConfig() {
	_, err := NewAgentWithConfig(&Config{
		ServerAddress: "http://localhost:80",
		TenantID:      "00000000-0000-4000-0000-000000000000",
		PrivateKey:    "./shellhub.key",
	})
	if err != nil {
		panic(err)
	}
}

func ExampleNewAgent() {
	_, err := NewAgent("http://localhost:80", "00000000-0000-4000-0000-000000000000", "./shellhub.key")
	if err != nil {
		panic(err)
	}
}

func TestLoadConfigFromEnv(t *testing.T) {
	envMock := new(env_mocks.Backend)
	envs.DefaultBackend = envMock

	type expected struct {
		cfg    *Config
		fields map[string]interface{}
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
					cfg := args.Get(1).(*Config)

					cfg.ServerAddress = "http://localhost"
					cfg.TenantID = ""
					cfg.PrivateKey = ""
					cfg.MaxRetryConnectionTimeout = 30
				})
			},
			expected: expected{
				cfg: nil,
				fields: map[string]interface{}{
					"TenantID":   "required",
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
					cfg := args.Get(1).(*Config)

					cfg.ServerAddress = ""
					cfg.TenantID = ""
					cfg.PrivateKey = ""
					cfg.MaxRetryConnectionTimeout = 30
				})
			},
			expected: expected{
				cfg: nil,
				fields: map[string]interface{}{
					"ServerAddress": "required",
					"TenantID":      "required",
					"PrivateKey":    "required",
				},
				err: validator.ErrStructureInvalid,
			},
		},
		{
			description: "success to load the environemental variables",
			requiredMocks: func() {
				envs := new(Config)

				envMock.On("Process", "SHELLHUB_", envs).Return(nil).Once().Run(func(args mock.Arguments) {
					cfg := args.Get(1).(*Config)

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

func TestNewAgentWithConfig(t *testing.T) {
	type expected struct {
		agent *Agent
		err   error
	}

	// NOTICE: configuration structure used by the successfully test.
	config := &Config{
		ServerAddress: "http://localhost",
		TenantID:      "1c462afa-e4b6-41a5-ba54-7236a1770466",
		PrivateKey:    "/tmp/shellhub.key",
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
			description: "fail when tenant is empty",
			config: &Config{
				ServerAddress: "http://localhost",
				TenantID:      "",
			},
			mode: new(HostMode),
			expected: expected{
				agent: nil,
				err:   ErrNewAgentWithConfigEmptyTenant,
			},
		},
		{
			description: "fail when private key is empty",
			config: &Config{
				ServerAddress: "http://localhost",
				TenantID:      "1c462afa-e4b6-41a5-ba54-7236a1770466",
				PrivateKey:    "",
			},
			mode: new(HostMode),
			expected: expected{
				agent: nil,
				err:   ErrNewAgentWithConfigEmptyPrivateKey,
			},
		},
		{
			description: "success to create agent with config",
			config:      config,
			mode:        new(HostMode),
			expected: expected{
				agent: &Agent{
					config: config,
				},
				err: nil,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			agent, err := NewAgentWithConfig(test.config)

			assert.Equal(t, test.expected.agent, agent)
			assert.ErrorIs(t, err, test.expected.err)
		})
	}
}

func TestAgent_GetInfo(t *testing.T) {
	clientMocks := new(client_mocks.Client)

	AgentVersion = "latest"

	type expected struct {
		info *models.Info
		err  error
	}

	agent := &Agent{
		cli: clientMocks,
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
