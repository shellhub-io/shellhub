package envs

import (
	"context"
	"os"

	"github.com/sethvargo/go-envconfig"
)

type envBackend struct{}

// envBackend is the default key/value store that reads from environment variables.
func (b *envBackend) Get(name string) string {
	return os.Getenv(name)
}

func (b *envBackend) Process(prefix string, spec any) error {
	return envconfig.ProcessWith(context.Background(), &envconfig.Config{
		Target: spec,
		Lookuper: envconfig.MultiLookuper(
			envconfig.PrefixLookuper(prefix, envconfig.OsLookuper()),
			envconfig.OsLookuper(),
		),
	})
}
