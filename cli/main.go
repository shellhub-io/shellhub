package main

import (
	"context"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/mongo"
	"github.com/shellhub-io/shellhub/api/store/pg"
	pgoptions "github.com/shellhub-io/shellhub/api/store/pg/options"
	"github.com/shellhub-io/shellhub/cli/cmd"
	"github.com/shellhub-io/shellhub/cli/services"
	"github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/loglevel"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type config struct {
	Database string `env:"DATABASE,default=mongo"`

	MongoURI string `env:"MONGO_URI,default=mongodb://mongo:27017/main"`

	// PostgresHost specifies the host for PostgreSQL.
	PostgresHost string `env:"POSTGRES_HOST,default=postgres"`
	// PostgresPort specifies the port for PostgreSQL.
	PostgresPort string `env:"POSTGRES_PORT,default=5432"`
	// PostgresUsername specifies the username for authenticate PostgreSQL.
	PostgresUsername string `env:"POSTGRES_USERNAME,default=admin"`
	// PostgresUser specifies the password for authenticate PostgreSQL.
	PostgresPassword string `env:"POSTGRES_PASSWORD,default=admin"`
	// PostgresDatabase especifica o nome do banco de dados PostgreSQL a ser utilizado.
	PostgresDatabase string `env:"POSTGRES_DATABASE,default=main"`

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

	log.Info("Connecting to Redis")

	cache, err := cache.NewRedisCache(cfg.RedisURI, 0)
	if err != nil {
		log.Fatal(err)
	}

	log.Info("Connected to Redis")

	log.Trace("Connecting to MongoDB")

	var store store.Store
	switch cfg.Database {
	case "mongo":
		store, err = mongo.NewStore(ctx, cfg.MongoURI, cache)
	case "postgres":
		uri := pg.URI(cfg.PostgresHost, cfg.PostgresPort, cfg.PostgresUsername, cfg.PostgresPassword, cfg.PostgresDatabase)
		store, err = pg.New(ctx, uri, pgoptions.Log("INFO", true)) // TODO: Log envs
	default:
		log.WithField("database", cfg.Database).Fatal("invalid database")
	}

	if err != nil {
		log.WithError(err).Fatal("failed to create the store")
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
