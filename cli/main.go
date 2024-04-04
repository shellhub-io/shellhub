package main

import (
	"context"

	"github.com/shellhub-io/shellhub/api/store/mongo"
	"github.com/shellhub-io/shellhub/cli/cmd"
	"github.com/shellhub-io/shellhub/cli/services"
	storecache "github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/loglevel"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	mgo "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
)

type config struct {
	MongoURI string `env:"MONGO_URI,default=mongodb://mongo:27017/main"`
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

	cache, err := storecache.NewRedisCache(cfg.RedisURI)
	if err != nil {
		log.WithError(err).Error("Failed to configure redis store cache")
	}

	log.Info("Connected to Redis")

	log.Trace("Connecting to MongoDB")

	_, db, err := mongo.Connect(ctx, cfg.MongoURI)
	if err != nil {
		log.
			WithError(err).
			Fatal("unable to connect to MongoDB")
	}

	store, err := mongo.NewStore(ctx, db, cache)
	if err != nil {
		log.
			WithError(err).
			Fatal("failed to create the store")
	}

	service := services.NewService(store)

	rootCmd := &cobra.Command{Use: "cli"}

	rootCmd.AddCommand(cmd.UserCommands(service))
	rootCmd.AddCommand(cmd.NamespaceCommands(service))
	cmd.DeprecatedCommands(rootCmd, service)

	if err := rootCmd.Execute(); err != nil {
		log.Error(err)
	}
}
