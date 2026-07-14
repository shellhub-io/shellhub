package envstest

import (
	"testing"

	"github.com/shellhub-io/shellhub/pkg/envs"
)

type mapBackend struct {
	values map[string]string
}

func (m *mapBackend) Get(key string) string {
	return m.values[key]
}

func (m *mapBackend) Process(string, interface{}) error {
	return nil
}

// SetEdition swaps envs.DefaultBackend with a simple map backend that returns
// the given edition for SHELLHUB_EDITION, and restores the original on cleanup.
func SetEdition(t *testing.T, edition envs.Edition) {
	t.Helper()

	SetRawEdition(t, string(edition))
}

// SetRawEdition is like SetEdition but accepts an arbitrary string, including
// invalid or un-normalized values. Use it to test CurrentEdition's
// trim/lowercase/validation logic.
func SetRawEdition(t *testing.T, raw string) {
	t.Helper()

	prev := envs.DefaultBackend
	t.Cleanup(func() { envs.DefaultBackend = prev })

	envs.DefaultBackend = &mapBackend{values: map[string]string{
		"SHELLHUB_EDITION": raw,
	}}
}
