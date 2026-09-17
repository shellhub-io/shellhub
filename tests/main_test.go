package main

import (
	"os"
	"testing"

	"github.com/shellhub-io/shellhub/tests/environment"
	log "github.com/sirupsen/logrus"
)

func TestMain(m *testing.M) {
	_ = os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")

	if err := environment.GenerateKeys(".."); err != nil {
		log.WithError(err).Error("failed to generate the ShellHub keys")

		os.Exit(1)
	}

	os.Exit(m.Run())
}
