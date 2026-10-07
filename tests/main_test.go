package main

import (
	"context"
	"os"
	"testing"

	"github.com/shellhub-io/shellhub/tests/environment"
	log "github.com/sirupsen/logrus"
)

const licenseIssuerKeyPath = ".stack/license-issuer.pem"

var run *environment.Run

func TestMain(m *testing.M) {
	os.Exit(runSuite(m))
}

func runSuite(m *testing.M) int {
	_ = os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")

	ctx := context.Background()

	for _, edition := range []environment.Edition{environment.EditionCommunity, environment.EditionEnterprise} {
		if err := environment.BundleOpenAPI(ctx, edition, ".."); err != nil {
			log.WithError(err).WithField("edition", edition).Error("failed to bundle the OpenAPI schema")

			return 1
		}
	}

	if err := environment.GenerateKeys(".."); err != nil {
		log.WithError(err).Error("failed to generate the ShellHub keys")

		return 1
	}

	issuer, err := environment.LoadLicenseIssuer(licenseIssuerKeyPath)
	if err != nil {
		log.WithError(err).Error("failed to load the license issuer")

		return 1
	}

	if run, err = environment.StartRun(ctx, environment.IssuingLicenses(issuer)); err != nil {
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
