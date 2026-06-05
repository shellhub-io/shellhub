package main

import (
	"reflect"
	"testing"
)

// TestConfigIsPostgresOnly verifies that the config struct has been
// rewired to postgres-only. It must not contain the Database switch field
// or the MongoURI field that belonged to the old mongo/switch-based construction.
func TestConfigIsPostgresOnly(t *testing.T) {
	t.Parallel()

	cfg := config{}
	typ := reflect.TypeOf(cfg)

	unwanted := []string{"Database", "MongoURI"}
	for _, field := range unwanted {
		if _, ok := typ.FieldByName(field); ok {
			t.Errorf("config struct must not contain field %q after pg-only rewire", field)
		}
	}

	required := []string{
		"PostgresHost",
		"PostgresPort",
		"PostgresUsername",
		"PostgresPassword",
		"PostgresDatabase",
		"PostgresLogLevel",
		"PostgresLogVerbose",
	}
	for _, field := range required {
		if _, ok := typ.FieldByName(field); !ok {
			t.Errorf("config struct is missing required pg field %q", field)
		}
	}
}
