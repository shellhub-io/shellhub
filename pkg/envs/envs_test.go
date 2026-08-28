package envs_test

import (
	"testing"

	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/envs/envstest"
	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
)

func TestCurrentEdition(t *testing.T) {
	tests := []struct {
		description string
		envValue    string
		expected    envs.Edition
	}{
		{
			description: "unset defaults to community",
			envValue:    "",
			expected:    envs.Community,
		},
		{
			description: "explicit community",
			envValue:    "community",
			expected:    envs.Community,
		},
		{
			description: "explicit enterprise",
			envValue:    "enterprise",
			expected:    envs.Enterprise,
		},
		{
			description: "explicit cloud",
			envValue:    "cloud",
			expected:    envs.Cloud,
		},
		{
			description: "uppercase is normalized",
			envValue:    "CLOUD",
			expected:    envs.Cloud,
		},
		{
			description: "mixed case is normalized",
			envValue:    "Enterprise",
			expected:    envs.Enterprise,
		},
		{
			description: "whitespace is trimmed",
			envValue:    "  cloud  ",
			expected:    envs.Cloud,
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			envstest.SetRawEdition(t, tt.envValue)

			assert.Equal(t, tt.expected, envs.CurrentEdition())
		})
	}
}

func TestCurrentEdition_invalid_panics(t *testing.T) {
	envstest.SetRawEdition(t, "invalid")

	assert.Panics(t, func() { envs.CurrentEdition() })
}

func TestResolveEdition(t *testing.T) {
	tests := []struct {
		description string
		envValue    string
		expected    envs.Edition
		expectErr   bool
	}{
		{
			description: "unset defaults to community",
			envValue:    "",
			expected:    envs.Community,
		},
		{
			description: "normalizes case and whitespace",
			envValue:    "  Cloud  ",
			expected:    envs.Cloud,
		},
		{
			description: "invalid returns an error instead of panicking",
			envValue:    "invalid",
			expectErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			envstest.SetRawEdition(t, tt.envValue)

			edition, err := envs.ResolveEdition()
			if tt.expectErr {
				assert.Error(t, err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, edition)
		})
	}
}

func TestEditionPredicates(t *testing.T) {
	tests := []struct {
		edition             envs.Edition
		isCommunity         bool
		isEnterprise        bool
		isCloud             bool
		isEnterpriseOrCloud bool
	}{
		{envs.Community, true, false, false, false},
		{envs.Enterprise, false, true, false, true},
		{envs.Cloud, false, false, true, true},
	}

	for _, tt := range tests {
		t.Run(string(tt.edition), func(t *testing.T) {
			envstest.SetEdition(t, tt.edition)

			assert.Equal(t, tt.isCommunity, envs.IsCommunity())
			assert.Equal(t, tt.isEnterprise, envs.IsEnterprise())
			assert.Equal(t, tt.isCloud, envs.IsCloud())
			assert.Equal(t, tt.isEnterpriseOrCloud, envs.IsEnterpriseOrCloud())
		})
	}
}

func TestParseWithPrefix_with_default(t *testing.T) {
	type Envs struct {
		RedisURI string `env:"REDIS_URI,default=redis://redis:6379/default"`
		MongoURI string `env:"MONGO_URI,default=mongodb://mongo:27017/default"`
	}

	type Expected struct {
		Envs  *Envs
		Error error
	}

	tests := []struct {
		description string
		prefix      string
		setup       func(t *testing.T)
		expected    Expected
	}{
		{
			description: "parse envs with prefix empty",
			prefix:      "",
			setup: func(t *testing.T) {
				t.Helper()

				t.Setenv("REDIS_URI", "redis://redis:6379/empty")
				t.Setenv("MONGO_URI", "mongodb://mongo:27017/empty")
			},
			expected: Expected{
				Envs: &Envs{
					RedisURI: "redis://redis:6379/empty",
					MongoURI: "mongodb://mongo:27017/empty",
				},
				Error: nil,
			},
		},
		{
			description: "parse envs with one prefix and an empty",
			prefix:      "FOO_",
			setup: func(t *testing.T) {
				t.Helper()

				t.Setenv("FOO_REDIS_URI", "redis://redis:6379/foo")
				t.Setenv("REDIS_URI", "redis://redis:6379/empty")
				t.Setenv("MONGO_URI", "mongodb://mongo:27017/empty")
			},
			expected: Expected{
				Envs: &Envs{
					RedisURI: "redis://redis:6379/foo",
					MongoURI: "mongodb://mongo:27017/empty",
				},
				Error: nil,
			},
		},
		{
			description: "parse envs with one prefix",
			prefix:      "BAR_",
			setup: func(t *testing.T) {
				t.Helper()

				t.Setenv("FOO_REDIS_URI", "redis://redis:6379/foo")
				t.Setenv("BAR_REDIS_URI", "redis://redis:6379/bar")
				t.Setenv("FOO_MONGO_URI", "mongodb://mongo:27017/foo")
				t.Setenv("BAR_MONGO_URI", "mongodb://mongo:27017/bar")
			},
			expected: Expected{
				Envs: &Envs{
					RedisURI: "redis://redis:6379/bar",
					MongoURI: "mongodb://mongo:27017/bar",
				},
				Error: nil,
			},
		},
		{
			description: "parse envs with one prefix and default",
			prefix:      "FOO_",
			setup: func(t *testing.T) {
				t.Helper()

				t.Setenv("FOO_REDIS_URI", "redis://redis:6379/foo")
				t.Setenv("BAR_REDIS_URI", "redis://redis:6379/bar")
				t.Setenv("BAR_MONGO_URI", "mongodb://mongo:27017/bar")
			},
			expected: Expected{
				Envs: &Envs{
					RedisURI: "redis://redis:6379/foo",
					MongoURI: "mongodb://mongo:27017/default",
				},
				Error: nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			tt.setup(t)

			result, err := envs.ParseWithPrefix[Envs](tt.prefix)
			assert.Equal(t, tt.expected.Envs, result)
			assert.ErrorIs(t, err, tt.expected.Error)
		})
	}
}

func TestParseWithPrefix_with_required(t *testing.T) {
	type Envs struct {
		RedisURI string `env:"REDIS_URI,required"`
		MongoURI string `env:"MONGO_URI,required"`
	}

	type Expected struct {
		Envs  *Envs
		Error error
	}

	tests := []struct {
		description string
		prefix      string
		setup       func(t *testing.T)
		expected    Expected
	}{
		{
			description: "parse envs with a prefix and no prefixed",
			prefix:      "FOO_",
			setup: func(t *testing.T) {
				t.Helper()

				t.Setenv("FOO_REDIS_URI", "redis://redis:6379/foo")
				t.Setenv("MONGO_URI", "mongodb://mongo:27017/empty")
			},
			expected: Expected{
				Envs: &Envs{
					RedisURI: "redis://redis:6379/foo",
					MongoURI: "mongodb://mongo:27017/empty",
				},
				Error: nil,
			},
		},
		{
			description: "parse envs with a prefix and no prefixed",
			prefix:      "FOO_",
			setup: func(t *testing.T) {
				t.Helper()

				t.Setenv("REDIS_URI", "redis://redis:6379/empty")
				t.Setenv("MONGO_URI", "mongodb://mongo:27017/empty")
			},
			expected: Expected{
				Envs: &Envs{
					RedisURI: "redis://redis:6379/empty",
					MongoURI: "mongodb://mongo:27017/empty",
				},
				Error: nil,
			},
		},
		{
			description: "fails to parse when two different prefixes",
			prefix:      "FOO_",
			setup: func(t *testing.T) {
				t.Helper()

				t.Setenv("FOO_REDIS_URI", "redis://redis:6379/foo")
				t.Setenv("BAR_MONGO_URI", "mongodb://mongo:27017/empty")
			},
			expected: Expected{
				Envs:  nil,
				Error: envs.ErrParseWithPrefix,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			tt.setup(t)

			result, err := envs.ParseWithPrefix[Envs](tt.prefix)
			assert.Equal(t, tt.expected.Envs, result)
			assert.ErrorIs(t, err, tt.expected.Error)
		})
	}
}

func TestParse_with_default(t *testing.T) {
	type Envs struct {
		RedisURI string `env:"REDIS_URI,default=redis://redis:6379/default"`
		MongoURI string `env:"MONGO_URI,default=mongodb://mongo:27017/default"`
	}

	type Expected struct {
		Envs  *Envs
		Error error
	}

	tests := []struct {
		description string
		setup       func(t *testing.T)
		expected    Expected
	}{
		{
			description: "parse envs",
			setup: func(t *testing.T) {
				t.Helper()

				t.Setenv("REDIS_URI", "redis://redis:6379/test")
				t.Setenv("MONGO_URI", "mongodb://mongo:27017/test")
			},
			expected: Expected{
				Envs: &Envs{
					RedisURI: "redis://redis:6379/test",
					MongoURI: "mongodb://mongo:27017/test",
				},
				Error: nil,
			},
		},
		{
			description: "parse envs with one set and one default",
			setup: func(t *testing.T) {
				t.Helper()

				t.Setenv("REDIS_URI", "redis://redis:6379/test")
			},
			expected: Expected{
				Envs: &Envs{
					RedisURI: "redis://redis:6379/test",
					MongoURI: "mongodb://mongo:27017/default",
				},
				Error: nil,
			},
		},
		{
			description: "parse envs with all default",
			setup: func(t *testing.T) {
				t.Helper()
			},
			expected: Expected{
				Envs: &Envs{
					RedisURI: "redis://redis:6379/default",
					MongoURI: "mongodb://mongo:27017/default",
				},
				Error: nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			tt.setup(t)

			result, err := envs.Parse[Envs]()
			assert.Equal(t, tt.expected.Envs, result)
			assert.ErrorIs(t, err, tt.expected.Error)
		})
	}
}

func TestParse_with_required(t *testing.T) {
	type Envs struct {
		RedisURI string `env:"REDIS_URI,required"`
		MongoURI string `env:"MONGO_URI,required"`
	}

	type Expected struct {
		Envs  *Envs
		Error error
	}

	tests := []struct {
		description string
		setup       func(t *testing.T)
		expected    Expected
	}{
		{
			description: "parse envs",
			setup: func(t *testing.T) {
				t.Helper()

				t.Setenv("REDIS_URI", "redis://redis:6379/test")
				t.Setenv("MONGO_URI", "mongodb://mongo:27017/test")
			},
			expected: Expected{
				Envs: &Envs{
					RedisURI: "redis://redis:6379/test",
					MongoURI: "mongodb://mongo:27017/test",
				},
				Error: nil,
			},
		},
		{
			description: "fail to parse envs when one env is missing",
			setup: func(t *testing.T) {
				t.Helper()

				t.Setenv("REDIS_URI", "redis://redis:6379/test")
			},
			expected: Expected{
				Error: envs.ErrParse,
			},
		},
		{
			description: "fails to parse when all envs are missing",
			setup: func(t *testing.T) {
				t.Helper()
			},
			expected: Expected{
				Envs:  nil,
				Error: envs.ErrParse,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			tt.setup(t)

			result, err := envs.Parse[Envs]()
			assert.Equal(t, tt.expected.Envs, result)
			assert.ErrorIs(t, err, tt.expected.Error)
		})
	}
}
