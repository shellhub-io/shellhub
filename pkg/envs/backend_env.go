package envs

import (
	"os"

	"github.com/caarlos0/env/v10"
)

type envBackend struct{}

// envBackend is the default key/value store that reads from environment variables.
func (b *envBackend) Get(name string) string {
	return os.Getenv(name)
}

func (b *envBackend) Process(prefix string, spec interface{}) error {
	return env.ParseWithOptions(spec, env.Options{
		Prefix: prefix,
	})
}
