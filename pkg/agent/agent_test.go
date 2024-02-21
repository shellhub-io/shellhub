package agent

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/shellhub-io/shellhub/pkg/envs"
	env_mocks "github.com/shellhub-io/shellhub/pkg/envs/mocks"
	"github.com/shellhub-io/shellhub/pkg/validator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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
				})
			},
			expected: expected{
				cfg: &Config{
					ServerAddress: "http://localhost",
					TenantID:      "1c462afa-e4b6-41a5-ba54-7236a1770466",
					PrivateKey:    "/tmp/shellhub.key",
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
