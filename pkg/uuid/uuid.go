package uuid

import (
	uuid "github.com/satori/go.uuid" //nolint
)

// Uuid is an interface that can provide uuid related functionality which allows us to test uuid dependent code
type Uuid interface {
	Generate() string
}

// DefaultBackend is used to configure the defaultBackend
var DefaultBackend Uuid

// The init function will set the defaultBackend to the goUuid implementation
func init() {
	DefaultBackend = &goUuid{}
}

// Is responsible for calling method Generate of the defaultBackend
func Generate() string {
	return DefaultBackend.Generate()
}

type goUuid struct {
}

// This function is responsible for generating UUID v4 of the satori package
func (g *goUuid) Generate() string {
	return uuid.Must(uuid.NewV4(), nil).String()
}
