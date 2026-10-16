package main

import (
	"context"
	"flag"
	"os"
	"testing"

	"github.com/shellhub-io/shellhub/tests/environment"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

var run *environment.Run

func TestMain(m *testing.M) {
	os.Exit(runSuite(m))
}

func runSuite(m *testing.M) int {
	flag.Parse()

	_ = os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")

	ctx := context.Background()

	for _, edition := range []environment.Edition{environment.EditionCommunity, environment.EditionEnterprise, environment.EditionCloud} {
		if err := environment.BundleOpenAPI(ctx, edition, ".."); err != nil {
			log.WithError(err).WithField("edition", edition).Error("failed to bundle the OpenAPI schema")

			return 1
		}
	}

	if err := environment.GenerateKeys(".."); err != nil {
		log.WithError(err).Error("failed to generate the ShellHub keys")

		return 1
	}

	issuer, err := environment.LoadLicenseIssuer("")
	if err != nil {
		log.WithError(err).Error("failed to load the license issuer")

		return 1
	}

	if run, err = environment.StartRun(ctx, issuer); err != nil {
		log.WithError(err).Error("failed to start the e2e run")

		return 1
	}

	defer func() {
		if err := run.Close(ctx); err != nil {
			log.WithError(err).Warn("the e2e run left objects behind for the next run to sweep")
		}
	}()

	if flag.Lookup("test.list").Value.String() == "" {
		if err := buildImages(ctx); err != nil {
			log.WithError(err).Error("failed to build the e2e images")

			return 1
		}
	}

	return m.Run()
}

func buildImages(ctx context.Context) error {
	group, ctx := errgroup.WithContext(ctx)

	group.Go(func() error { return environment.BuildStackImages(ctx, run, "") })
	group.Go(func() error {
		_, err := buildAgentImage(ctx, "")

		return err
	})

	return group.Wait()
}
