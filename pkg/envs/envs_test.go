package envs_test

import (
	"testing"

	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseEdition(t *testing.T) {
	t.Parallel()

	tests := []struct {
		description string
		raw         string
		expected    envs.Edition
		expectErr   bool
	}{
		{description: "empty is community", raw: "", expected: envs.Community},
		{description: "explicit community", raw: "community", expected: envs.Community},
		{description: "explicit enterprise", raw: "enterprise", expected: envs.Enterprise},
		{description: "explicit cloud", raw: "cloud", expected: envs.Cloud},
		{description: "uppercase is normalized", raw: "CLOUD", expected: envs.Cloud},
		{description: "whitespace is trimmed", raw: "  cloud  ", expected: envs.Cloud},
		{description: "unknown value is an error", raw: "invalid", expectErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			t.Parallel()

			edition, err := envs.ParseEdition(tt.raw)
			if tt.expectErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, edition)
		})
	}
}

func TestEditionPredicates(t *testing.T) {
	t.Parallel()

	tests := []struct {
		description         string
		edition             envs.Edition
		isCommunity         bool
		isEnterprise        bool
		isCloud             bool
		isEnterpriseOrCloud bool
	}{
		{"zero value", "", true, false, false, false},
		{"community", envs.Community, true, false, false, false},
		{"enterprise", envs.Enterprise, false, true, false, true},
		{"cloud", envs.Cloud, false, false, true, true},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.isCommunity, tt.edition.IsCommunity())
			assert.Equal(t, tt.isEnterprise, tt.edition.IsEnterprise())
			assert.Equal(t, tt.isCloud, tt.edition.IsCloud())
			assert.Equal(t, tt.isEnterpriseOrCloud, tt.edition.IsEnterpriseOrCloud())
		})
	}
}

func TestParse_edition(t *testing.T) {
	t.Parallel()

	type Envs struct {
		Edition envs.Edition `env:"SHELLHUB_EDITION"`
	}

	t.Run("unset is the zero value", func(t *testing.T) {
		t.Parallel()

		result, err := envs.Parse[Envs](envs.Map(nil))
		require.NoError(t, err)
		assert.True(t, result.Edition.IsCommunity())
	})

	t.Run("a valid value is normalized", func(t *testing.T) {
		t.Parallel()

		result, err := envs.Parse[Envs](envs.Map(map[string]string{"SHELLHUB_EDITION": " Cloud "}))
		require.NoError(t, err)
		assert.Equal(t, envs.Cloud, result.Edition)
	})

	t.Run("an invalid value fails the parse", func(t *testing.T) {
		t.Parallel()

		_, err := envs.Parse[Envs](envs.Map(map[string]string{"SHELLHUB_EDITION": "invalid"}))
		assert.ErrorIs(t, err, envs.ErrParse)
	})
}

func TestParseWithPrefix_with_default(t *testing.T) {
	t.Parallel()

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
		values      map[string]string
		expected    Expected
	}{
		{
			description: "parse envs with prefix empty",
			prefix:      "",
			values: map[string]string{
				"REDIS_URI": "redis://redis:6379/empty",
				"MONGO_URI": "mongodb://mongo:27017/empty",
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
			values: map[string]string{
				"FOO_REDIS_URI": "redis://redis:6379/foo",
				"REDIS_URI":     "redis://redis:6379/empty",
				"MONGO_URI":     "mongodb://mongo:27017/empty",
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
			values: map[string]string{
				"FOO_REDIS_URI": "redis://redis:6379/foo",
				"BAR_REDIS_URI": "redis://redis:6379/bar",
				"FOO_MONGO_URI": "mongodb://mongo:27017/foo",
				"BAR_MONGO_URI": "mongodb://mongo:27017/bar",
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
			values: map[string]string{
				"FOO_REDIS_URI": "redis://redis:6379/foo",
				"BAR_REDIS_URI": "redis://redis:6379/bar",
				"BAR_MONGO_URI": "mongodb://mongo:27017/bar",
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
			t.Parallel()

			result, err := envs.ParseWithPrefix[Envs](envs.Map(tt.values), tt.prefix)
			assert.Equal(t, tt.expected.Envs, result)
			assert.ErrorIs(t, err, tt.expected.Error)
		})
	}
}

func TestParseWithPrefix_with_required(t *testing.T) {
	t.Parallel()

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
		values      map[string]string
		expected    Expected
	}{
		{
			description: "parse envs with a prefix and no prefixed",
			prefix:      "FOO_",
			values: map[string]string{
				"FOO_REDIS_URI": "redis://redis:6379/foo",
				"MONGO_URI":     "mongodb://mongo:27017/empty",
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
			description: "parse envs with a prefix and only unprefixed",
			prefix:      "FOO_",
			values: map[string]string{
				"REDIS_URI": "redis://redis:6379/empty",
				"MONGO_URI": "mongodb://mongo:27017/empty",
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
			values: map[string]string{
				"FOO_REDIS_URI": "redis://redis:6379/foo",
				"BAR_MONGO_URI": "mongodb://mongo:27017/empty",
			},
			expected: Expected{
				Envs:  nil,
				Error: envs.ErrParseWithPrefix,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			t.Parallel()

			result, err := envs.ParseWithPrefix[Envs](envs.Map(tt.values), tt.prefix)
			assert.Equal(t, tt.expected.Envs, result)
			assert.ErrorIs(t, err, tt.expected.Error)
		})
	}
}

func TestParseWithPrefix_map_matches_process(t *testing.T) {
	type Envs struct {
		RedisURI string `env:"REDIS_URI"`
	}

	values := map[string]string{
		"FOO_REDIS_URI": "redis://redis:6379/prefixed",
		"REDIS_URI":     "redis://redis:6379/unprefixed",
	}

	for key, value := range values {
		t.Setenv(key, value)
	}

	fromMap, err := envs.ParseWithPrefix[Envs](envs.Map(values), "FOO_")
	require.NoError(t, err)

	fromProcess, err := envs.ParseWithPrefix[Envs](envs.OS(), "FOO_")
	require.NoError(t, err)

	assert.Equal(t, "redis://redis:6379/prefixed", fromMap.RedisURI)
	assert.Equal(t, fromProcess, fromMap)
}

func TestParse_with_default(t *testing.T) {
	t.Parallel()

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
		values      map[string]string
		expected    Expected
	}{
		{
			description: "parse envs",
			values: map[string]string{
				"REDIS_URI": "redis://redis:6379/test",
				"MONGO_URI": "mongodb://mongo:27017/test",
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
			values: map[string]string{
				"REDIS_URI": "redis://redis:6379/test",
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
			values:      map[string]string{},
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
			t.Parallel()

			result, err := envs.Parse[Envs](envs.Map(tt.values))
			assert.Equal(t, tt.expected.Envs, result)
			assert.ErrorIs(t, err, tt.expected.Error)
		})
	}
}

func TestParse_with_required(t *testing.T) {
	t.Parallel()

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
		values      map[string]string
		expected    Expected
	}{
		{
			description: "parse envs",
			values: map[string]string{
				"REDIS_URI": "redis://redis:6379/test",
				"MONGO_URI": "mongodb://mongo:27017/test",
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
			values: map[string]string{
				"REDIS_URI": "redis://redis:6379/test",
			},
			expected: Expected{
				Error: envs.ErrParse,
			},
		},
		{
			description: "fails to parse when all envs are missing",
			values:      map[string]string{},
			expected: Expected{
				Envs:  nil,
				Error: envs.ErrParse,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			t.Parallel()

			result, err := envs.Parse[Envs](envs.Map(tt.values))
			assert.Equal(t, tt.expected.Envs, result)
			assert.ErrorIs(t, err, tt.expected.Error)
		})
	}
}
