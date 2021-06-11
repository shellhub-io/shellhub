package main

import (
	"context"
	"fmt"

	"github.com/kelseyhightower/envconfig"
	storecache "github.com/shellhub-io/shellhub/api/store/cache"
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
		log.Error((err.Error()))
	}

	client, err := mgo.Connect(context.TODO(), options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		log.Error(err)
	}

	var cache storecache.Cache

	if cfg.StoreCache {
		cache, err = storecache.NewRedisCache(cfg.RedisURI)
		log.Error(err)
	} else {
		cache = storecache.NewNullCache()
	}

	svc := NewService(mongo.NewStore(client.Database("main"), cache))

	var rootCmd = &cobra.Command{Use: "cli"}
	rootCmd.AddCommand(&cobra.Command{
		Use:   "add-user",
		Short: "Usage: <username> <password> <email>",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			username, err := svc.UserCreate(Arguments{
				Username: args[0],
				Password: args[1],
				Email:    args[2],
			})
			if err != nil {
				return err
			}

			fmt.Println("User added: ", username) //nolint:forbidigo

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

				ns, err := svc.NamespaceCreate(Arguments{
					Namespace: args[0],
					Username:  args[1],
					TenantID:  args[2],
				})
				if err != nil {
					return err
				}

				fmt.Println("Namespace added: ", ns.Name) //nolint:forbidigo
				fmt.Println("Owner: ", ns.Owner)          //nolint:forbidigo
				fmt.Println("Tenant ID: ", ns.TenantID)   //nolint:forbidigo

				return nil
			},
		},

		&cobra.Command{
			Use:   "add-user-namespace",
			Short: "Usage: <username> <namespace>",
			Args:  cobra.ExactArgs(2),
			RunE: func(cmd *cobra.Command, args []string) error {
				ns, err := svc.NamespaceAddMember(Arguments{
					Username:  args[0],
					Namespace: args[1],
				})
				if err != nil {
					return err
				}

				fmt.Println("User: ", ns.Owner)              //nolint:forbidigo
				fmt.Println("added to namespace: ", ns.Name) //nolint:forbidigo

				return nil
			},
		},

		&cobra.Command{
			Use:   "del-namespace",
			Short: "Usage: <namespace>",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				if err := svc.NamespaceDelete(Arguments{
					Namespace: args[0],
				}); err != nil {
					return err
				}

				fmt.Println("Namespace deleted") //nolint:forbidigo

				return nil
			},
		},

		&cobra.Command{
			Use:   "del-user",
			Short: "Usage: <username>",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				if err := svc.UserDelete(Arguments{
					Username: args[0],
				}); err != nil {
					return err
				}

				fmt.Println("User deleted") //nolint:forbidigo

				return nil
			},
		},

		&cobra.Command{
			Use:   "del-user-namespace",
			Short: "Usage <username> <namespace>",
			Args:  cobra.ExactArgs(2),
			RunE: func(cmd *cobra.Command, args []string) error {
				ns, err := svc.NamespaceRemoveMember(Arguments{
					Username:  args[0],
					Namespace: args[1],
				})
				if err != nil {
					return err
				}

				fmt.Println("User: ", ns.Owner)                  //nolint:forbidigo
				fmt.Println("removed from namespace: ", ns.Name) //nolint:forbidigo

				return nil
			},
		},

		&cobra.Command{
			Use:   "reset-user-password",
			Short: "Usage: <username> <password>",
			Args:  cobra.ExactArgs(2),
			RunE: func(cmd *cobra.Command, args []string) error {
				if err := svc.UserUpdate(Arguments{
					Username: args[0],
					Password: args[1],
				}); err != nil {
					return err
				}

				fmt.Println("Password changed") //nolint:forbidigo

				return nil
			},
		})

	if err := rootCmd.Execute(); err != nil {
		log.Error(err)
	}
}
