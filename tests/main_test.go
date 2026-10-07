package main

import (
	"context"
	"os"
	"testing"

	"github.com/shellhub-io/shellhub/tests/environment"
	log "github.com/sirupsen/logrus"
)

var run *environment.Run

func TestMain(m *testing.M) {
	os.Exit(runSuite(m))
}

func runSuite(m *testing.M) int {
	_ = os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")

	ctx := context.Background()

	if err := environment.BundleOpenAPI(ctx, environment.EditionCommunity, ".."); err != nil {
		log.WithError(err).Error("failed to bundle the OpenAPI schema")

		return 1
	}

	if err := environment.GenerateKeys(".."); err != nil {
		log.WithError(err).Error("failed to generate the ShellHub keys")

		return 1
	}

	var err error
	if run, err = environment.StartRun(ctx); err != nil {
		log.WithError(err).Error("failed to start the e2e run")

		return 1
	}

	defer func() {
		if err := run.Close(ctx); err != nil {
			log.WithError(err).Warn("the e2e run left objects behind for the next run to sweep")
		}
	}()

	return m.Run()
}
