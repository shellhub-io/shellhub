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
