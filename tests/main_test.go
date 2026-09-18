package main

import (
	"context"
	"os"
	"testing"

	"github.com/shellhub-io/shellhub/tests/environment"
	log "github.com/sirupsen/logrus"
)

func TestMain(m *testing.M) {
	_ = os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")

	if err := environment.BundleOpenAPI(context.Background(), environment.EditionCommunity, ".."); err != nil {
		log.WithError(err).Error("failed to bundle the OpenAPI schema")

		os.Exit(1)
	}

	if err := environment.GenerateKeys(".."); err != nil {
		log.WithError(err).Error("failed to generate the ShellHub keys")

		os.Exit(1)
	}

	os.Exit(m.Run())
}
