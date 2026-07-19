package envs_test

import (
	"os"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/envs/envstest"
	"github.com/stretchr/testify/assert"
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

			assert.NoError(t, err)
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
		before      func()
		after       func()
		expected    Expected
	}{
		{
			description: "parse envs with prefix empty",
			prefix:      "",
			before: func() {
				os.Setenv("REDIS_URI", "redis://redis:6379/empty")
				os.Setenv("MONGO_URI", "mongodb://mongo:27017/empty")
			},
			after: func() {
				os.Unsetenv("REDIS_URI")
				os.Unsetenv("MONGO_URI")
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
			before: func() {
				os.Setenv("FOO_REDIS_URI", "redis://redis:6379/foo")
				os.Setenv("REDIS_URI", "redis://redis:6379/empty")
				os.Setenv("MONGO_URI", "mongodb://mongo:27017/empty")
			},
			after: func() {
				os.Unsetenv("FOO_REDIS_URI")
				os.Unsetenv("REDIS_URI")
				os.Unsetenv("MONGO_URI")
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
			before: func() {
				os.Setenv("FOO_REDIS_URI", "redis://redis:6379/foo")
				os.Setenv("BAR_REDIS_URI", "redis://redis:6379/bar")
				os.Setenv("FOO_MONGO_URI", "mongodb://mongo:27017/foo")
				os.Setenv("BAR_MONGO_URI", "mongodb://mongo:27017/bar")
			},
			after: func() {
				os.Unsetenv("FOO_REDIS_URI")
				os.Unsetenv("BAR_REDIS_URI")
				os.Unsetenv("FOO_MONGO_URI")
				os.Unsetenv("BAR_MONGO_URI")
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
			before: func() {
				os.Setenv("FOO_REDIS_URI", "redis://redis:6379/foo")
				os.Setenv("BAR_REDIS_URI", "redis://redis:6379/bar")
				os.Setenv("BAR_MONGO_URI", "mongodb://mongo:27017/bar")
			},
			after: func() {
				os.Unsetenv("FOO_REDIS_URI")
				os.Unsetenv("BAR_REDIS_URI")
				os.Unsetenv("BAR_MONGO_URI")
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
			tt.before()

			result, err := envs.ParseWithPrefix[Envs](tt.prefix)
			assert.Equal(t, tt.expected.Envs, result)
			assert.ErrorIs(t, err, tt.expected.Error)

			tt.after()
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
		before      func()
		after       func()
		expected    Expected
	}{
		{
			description: "parse envs with a prefix and no prefixed",
			prefix:      "FOO_",
			before: func() {
				os.Setenv("FOO_REDIS_URI", "redis://redis:6379/foo")
				os.Setenv("MONGO_URI", "mongodb://mongo:27017/empty")
			},
			after: func() {
				os.Unsetenv("FOO_REDIS_URI")
				os.Unsetenv("MONGO_URI")
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
			before: func() {
				os.Setenv("REDIS_URI", "redis://redis:6379/empty")
				os.Setenv("MONGO_URI", "mongodb://mongo:27017/empty")
			},
			after: func() {
				os.Unsetenv("FOO_REDIS_URI")
				os.Unsetenv("MONGO_URI")
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
			before: func() {
				os.Setenv("FOO_REDIS_URI", "redis://redis:6379/foo")
				os.Setenv("BAR_MONGO_URI", "mongodb://mongo:27017/empty")
			},
			after: func() {
				os.Unsetenv("FOO_REDIS_URI")
				os.Unsetenv("BAR_MONGO_URI")
			},
			expected: Expected{
				Envs:  nil,
				Error: envs.ErrParseWithPrefix,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			tt.before()

			result, err := envs.ParseWithPrefix[Envs](tt.prefix)
			assert.Equal(t, tt.expected.Envs, result)
			assert.ErrorIs(t, err, tt.expected.Error)

			tt.after()
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
		before      func()
		after       func()
		expected    Expected
	}{
		{
			description: "parse envs",
			before: func() {
				os.Setenv("REDIS_URI", "redis://redis:6379/test")
				os.Setenv("MONGO_URI", "mongodb://mongo:27017/test")
			},
			after: func() {
				os.Unsetenv("REDIS_URI")
				os.Unsetenv("MONGO_URI")
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
			before: func() {
				os.Setenv("REDIS_URI", "redis://redis:6379/test")
			},
			after: func() {
				os.Unsetenv("REDIS_URI")
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
			before:      func() {},
			after:       func() {},
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
			tt.before()

			result, err := envs.Parse[Envs]()
			assert.Equal(t, tt.expected.Envs, result)
			assert.ErrorIs(t, err, tt.expected.Error)

			tt.after()
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
		before      func()
		after       func()
		expected    Expected
	}{
		{
			description: "parse envs",
			before: func() {
				os.Setenv("REDIS_URI", "redis://redis:6379/test")
				os.Setenv("MONGO_URI", "mongodb://mongo:27017/test")
			},
			after: func() {
				os.Unsetenv("REDIS_URI")
				os.Unsetenv("MONGO_URI")
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
			before: func() {
				os.Setenv("REDIS_URI", "redis://redis:6379/test")
			},
			after: func() {
				os.Unsetenv("REDIS_URI")
			},
			expected: Expected{
				Error: envs.ErrParse,
			},
		},
		{
			description: "fails to parse when all envs are missing",
			before: func() {
			},
			after: func() {
			},
			expected: Expected{
				Envs:  nil,
				Error: envs.ErrParse,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			tt.before()

			result, err := envs.Parse[Envs]()
			assert.Equal(t, tt.expected.Envs, result)
			assert.ErrorIs(t, err, tt.expected.Error)

			tt.after()
		})
	}
}
