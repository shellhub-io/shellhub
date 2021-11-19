package main

import (
	"context"

	"github.com/kelseyhightower/envconfig"
	storecache "github.com/shellhub-io/shellhub/api/cache"
	"github.com/shellhub-io/shellhub/api/store/mongo"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	mgo "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type config struct {
	MongoURI   string `envconfig:"mongo_uri" default:"mongodb://mongo:27017"`
	RedisURI   string `envconfig:"redis_uri" default:"redis://redis:6379"`
	StoreCache bool   `envconfig:"store_cache" default:"false"`
}

func main() {
	var cfg config

	if err := envconfig.Process("cli", &cfg); err != nil {
		log.Error(err.Error())
	}

	client, err := mgo.Connect(context.Background(), options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		log.Error(err)
	}

	var cache storecache.Cache

	if cfg.StoreCache {
		cache, err = storecache.NewRedisCache(cfg.RedisURI)
		if err != nil {
			log.Error(err)
		}
	} else {
		cache = storecache.NewNullCache()
	}

	services := NewService(mongo.NewStore(client.Database("main"), cache))

	rootCmd := &cobra.Command{Use: "cli"}
	rootCmd.AddCommand(&cobra.Command{
		Use:   "add-user",
		Short: "Usage: <username> <password> <email>",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			user, err := services.UserCreate(args[0], args[1], args[2])
			if err != nil {
				return err
			}
			rootCmd.Println("User added!")
			rootCmd.Println("name:", user.Name)
			rootCmd.Println("username:", user.Username)
			rootCmd.Println("email:", user.Email)

			return nil
		},
	},

		&cobra.Command{
			Use:   "del-user",
			Short: "Usage: <username>",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				if err := services.UserDelete(args[0]); err != nil {
					return err
				}

				rootCmd.Println("User deleted")

				return nil
			},
		},

		&cobra.Command{
			Use:   "reset-user-password",
			Short: "Usage: <username> <password>",
			Args:  cobra.ExactArgs(2),
			RunE: func(cmd *cobra.Command, args []string) error {
				if err := services.UserUpdate(args[0], args[1]); err != nil {
					return err
				}

				rootCmd.Println("Password changed")

				return nil
			},
		},
		&cobra.Command{
			Use:   "add-namespace",
			Short: "Usage: <namespace> <owner>",
			Args:  cobra.RangeArgs(2, 3),
			RunE: func(cmd *cobra.Command, args []string) error {
				// Avoid panic when TenantID isn't provided.
				if len(args) == 2 {
					args = append(args, "")
				}

				namespace, err := services.NamespaceCreate(args[0], args[1], args[2])
				if err != nil {
					return err
				}

				rootCmd.Println("Namespace added:", namespace.Name)
				rootCmd.Println("Owner:", namespace.Owner)
				rootCmd.Println("Tenant ID:", namespace.TenantID)

				return nil
			},
		},

		&cobra.Command{
			Use:   "add-user-namespace",
			Short: "Usage: <username> <namespace>",
			Args:  cobra.ExactArgs(2),
			RunE: func(cmd *cobra.Command, args []string) error {
				ns, err := services.NamespaceAddMember(args[0], args[1])
				if err != nil {
					return err
				}

				rootCmd.Println("User:", ns.Owner)
				rootCmd.Println("added to namespace:", ns.Name)

				return nil
			},
		},

		&cobra.Command{
			Use:   "del-user-namespace",
			Short: "Usage <username> <namespace>",
			Args:  cobra.ExactArgs(2),
			RunE: func(cmd *cobra.Command, args []string) error {
				ns, err := services.NamespaceRemoveMember(args[0], args[1])
				if err != nil {
					return err
				}

				rootCmd.Println("User:", ns.Owner)
				rootCmd.Println("removed from namespace:", ns.Name)

				return nil
			},
		},

		&cobra.Command{
			Use:   "del-namespace",
			Short: "Usage: <namespace>",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				if err := services.NamespaceDelete(args[0]); err != nil {
					return err
				}

				rootCmd.Println("Namespace deleted")

				return nil
			},
		})

	if err := rootCmd.Execute(); err != nil {
		log.Error(err)
	}
}
