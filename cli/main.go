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
	cfg, err := envs.ParseWithPrefix[config]("CLI_")
	if err != nil {
		log.Error(err.Error())
	}

	connStr, err := connstring.ParseAndValidate(cfg.MongoURI)
	if err != nil {
		log.WithError(err).Fatal("Invalid Mongo URI format")
	}

	client, err := mgo.Connect(context.Background(), options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		log.Error(err)
	}

	cache, err := storecache.NewRedisCache(cfg.RedisURI, 0)
	if err != nil {
		log.Fatal(err)
	}

	service := services.NewService(mongo.NewStore(client.Database(connStr.Database), cache))

	rootCmd := &cobra.Command{Use: "cli"}

	rootCmd.AddCommand(cmd.UserCommands(service))
	rootCmd.AddCommand(cmd.NamespaceCommands(service))
	cmd.DeprecatedCommands(rootCmd, service)

	if err := rootCmd.Execute(); err != nil {
		log.Error(err)
	}
}
