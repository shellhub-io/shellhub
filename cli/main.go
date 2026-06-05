package main

import (
	"context"

	"github.com/shellhub-io/shellhub/api/store/pg"
	pgoptions "github.com/shellhub-io/shellhub/api/store/pg/options"
	"github.com/shellhub-io/shellhub/cli/cmd"
	"github.com/shellhub-io/shellhub/cli/services"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/loglevel"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type config struct {
	// PostgresHost specifies the host for PostgreSQL.
	PostgresHost string `env:"POSTGRES_HOST,default=postgres"`
	// PostgresPort specifies the port for PostgreSQL.
	PostgresPort string `env:"POSTGRES_PORT,default=5432"`
	// PostgresUsername specifies the username for authenticate PostgreSQL.
	PostgresUsername string `env:"POSTGRES_USERNAME,default=admin"`
	// PostgresPassword specifies the password for authenticate PostgreSQL.
	PostgresPassword string `env:"POSTGRES_PASSWORD,default=admin"`
	// PostgresDatabase specifies the name of the PostgreSQL database to use.
	PostgresDatabase string `env:"POSTGRES_DATABASE,default=main"`
	// PostgresLogLevel specifies the log level for PostgresSQL query logging.
	PostgresLogLevel string `env:"POSTGRES_LOG_LEVEL,default=INFO"`
	// PostgresLogVerbose specifies whether to enable verbose PostgreSQL query logging.
	PostgresLogVerbose bool `env:"POSTGRES_LOG_VERBOSE,default=false"`
}

func init() {
	loglevel.SetLogLevel()
}

func main() {
	ctx := context.Background()

	cfg, err := envs.ParseWithPrefix[config]("CLI_")
	if err != nil {
		log.WithError(err).Fatal("failed to parse config envs")
	}

	log.Trace("Connecting to PostgreSQL")

	uri := pg.URI(
		cfg.PostgresHost,
		cfg.PostgresPort,
		cfg.PostgresUsername,
		cfg.PostgresPassword,
		cfg.PostgresDatabase,
	)

	store, err := pg.New(ctx, uri, pgoptions.Log(cfg.PostgresLogLevel, cfg.PostgresLogVerbose))
	if err != nil {
		log.WithError(err).Fatal("failed to create the store")
	}

	service := services.NewService(store)

	rootCmd := &cobra.Command{Use: "cli"}
	rootCmd.AddCommand(
		cmd.UserCommands(service),
		cmd.NamespaceCommands(service),
	)

	if err := rootCmd.Execute(); err != nil {
		log.Error(err)
	}
}
