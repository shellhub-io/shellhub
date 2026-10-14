package envs

import (
	"context"

	"github.com/sethvargo/go-envconfig"
)

type lookuperBackend struct {
	lookuper envconfig.Lookuper
}

// OS returns the backend that reads the process environment. Only an entrypoint should call it,
// to parse its configuration once and pass the values down.
func OS() Backend {
	return &lookuperBackend{lookuper: envconfig.OsLookuper()}
}

// Map returns a backend that reads values instead of the process environment. Its Process
// resolves a prefixed key over the unprefixed one exactly as the backend [OS] returns does.
func Map(values map[string]string) Backend {
	return &lookuperBackend{lookuper: envconfig.MapLookuper(values)}
}

func (b *lookuperBackend) Get(key string) string {
	value, _ := b.lookuper.Lookup(key)

	return value
}

func (b *lookuperBackend) Process(prefix string, spec any) error {
	return envconfig.ProcessWith(context.Background(), &envconfig.Config{
		Target: spec,
		Lookuper: envconfig.MultiLookuper(
			envconfig.PrefixLookuper(prefix, b.lookuper),
			b.lookuper,
		),
	})
}
