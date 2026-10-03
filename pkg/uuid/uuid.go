package uuid

import (
	"github.com/google/uuid" //nolint
)

// UUID is an interface that can provide uuid related functionality which allows us to test uuid dependent code.
type UUID interface {
	Generate() string
}

// DefaultBackend is used to configure the defaultBackend.
var DefaultBackend UUID

func init() {
	DefaultBackend = &goUUID{}
}

// Generate returns a new UUID v4 from the package's backend. Use it rather than the uuid package
// directly, so a test can make identifiers deterministic.
func Generate() string {
	return DefaultBackend.Generate()
}

// Fix makes Generate return v until t and its subtests finish, then puts back the backend it
// replaced. Calls nest: an inner Fix gives back the outer one when its test ends. The backend is
// process-global, so a test that calls Fix must not run in parallel with one that generates UUIDs.
func Fix(t interface{ Cleanup(func()) }, v string) {
	Set(t, fixedUUID(v))
}

// Set serves Generate from u until t and its subtests finish, then puts back the backend it
// replaced. Use it when a test needs Generate to vary; Fix covers a single value. Like Fix, it swaps
// a process-global backend and is not safe under t.Parallel.
func Set(t interface{ Cleanup(func()) }, u UUID) {
	previous := DefaultBackend
	t.Cleanup(func() { DefaultBackend = previous })
	DefaultBackend = u
}

type fixedUUID string

func (f fixedUUID) Generate() string {
	return string(f)
}

type goUUID struct{}

// This function is responsible for generating UUID v4 of the google package.
func (g *goUUID) Generate() string {
	return uuid.NewString()
}

// Parse reads a UUID from its string form. It does not go through the backend: parsing has no
// behaviour worth substituting in a test.
func Parse(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}
