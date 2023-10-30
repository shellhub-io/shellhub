package envs

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

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

			envs, err := ParseWithPrefix[Envs](tt.prefix)
			assert.Equal(t, tt.expected.Envs, envs)
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
				Error: ErrParseWithPrefix,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			tt.before()

			envs, err := ParseWithPrefix[Envs](tt.prefix)
			assert.Equal(t, tt.expected.Envs, envs)
			assert.ErrorIs(t, err, tt.expected.Error)

			tt.after()
		})
	}
}
