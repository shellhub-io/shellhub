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

// The init function will set the defaultBackend to the goUuid implementation.
func init() {
	DefaultBackend = &goUUID{}
}

// Is responsible for calling method Generate of the defaultBackend.
func Generate() string {
	return DefaultBackend.Generate()
}

type goUUID struct{}

// This function is responsible for generating UUID v4 of the google package.
func (g *goUUID) Generate() string {
	return uuid.NewString()
}
