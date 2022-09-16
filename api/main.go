package main

import (
	"context"

	"github.com/kelseyhightower/envconfig"
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

	// Populates configuration based on environment variables prefixed with 'API_'
	var cfg config
	if err := envconfig.Process("api", &cfg); err != nil {
		logrus.WithError(err).Fatal("Failed to load environment variables")
	}

	ctx := context.WithValue(context.TODO(), "cfg", &cfg) //nolint:revive

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		logrus.Fatal(err)
	}
}
