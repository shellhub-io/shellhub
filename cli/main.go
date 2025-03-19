package main

import (
	"context"

	"github.com/shellhub-io/shellhub/api/store/pg"
	"github.com/shellhub-io/shellhub/cli/cmd"
	"github.com/shellhub-io/shellhub/cli/services"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/loglevel"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type config struct {
	MongoURI string `env:"MONGO_URI,default=mongodb://mongo:27017/main"`

	// PostgresHost specifies the host for PostgreSQL.
	PostgresHost string `env:"POSTGRES_HOST,default=postgres"`
	// PostgresPort specifies the port for PostgreSQL.
	PostgresPort string `env:"POSTGRES_PORT,default=5432"`
	// PostgresUser specifies the username for authenticate PostgreSQL.
	PostgresUser string `env:"POSTGRES_USER,default=admin"`
	// PostgresUser specifies the password for authenticate PostgreSQL.
	PostgresPassword string `env:"POSTGRES_PASSWORD,default=admin"`
	// PostgresDB especifica o nome do banco de dados PostgreSQL a ser utilizado.
	PostgresDB string `env:"POSTGRES_DB,default=main"`

	RedisURI string `env:"REDIS_URI,default=redis://redis:6379"`
}

func init() {
	loglevel.SetLogLevel()
}

func main() {
	ctx := context.Background()

	cfg, err := envs.ParseWithPrefix[config]("CLI_")
	if err != nil {
		log.Error(err.Error())
	}

	uri := pg.URI(cfg.PostgresHost, cfg.PostgresPort, cfg.PostgresUser, cfg.PostgresPassword, cfg.PostgresDB)
	store, err := pg.New(ctx, uri)
	if err != nil {
		log.
			WithError(err).
			Fatal("failed to create the store")
	}

	service := services.NewService(store)

	rootCmd := &cobra.Command{Use: "cli"}
	rootCmd.AddCommand(cmd.UserCommands(service))
	rootCmd.AddCommand(cmd.NamespaceCommands(service))
	// WARN: this is deprecated and will be removed soon
	cmd.DeprecatedCommands(rootCmd, service)

	if err := rootCmd.Execute(); err != nil {
		log.Error(err)
	}
}
