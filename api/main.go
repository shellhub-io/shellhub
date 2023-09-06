package main

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/loglevel"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	loglevel.SetLogLevel()
}

func main() {
	rootCmd := &cobra.Command{Use: "api"}

	rootCmd.AddCommand(serverCmd)

	// Populates configuration based on environment variables prefixed with 'API_'.
	cfg, err := envs.ParseWithPrefix[config]("api")
	if err != nil {
		logrus.WithError(err).Fatal("Failed to load environment variables")
	}

	ctx := context.WithValue(context.TODO(), "cfg", cfg) //nolint:revive

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		logrus.Fatal(err)
	}
}
