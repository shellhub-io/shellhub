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
	MongoUri   string `envconfig:"mongo_uri" default:"mongodb://mongo:27017"`
	StoreCache bool   `envconfig:"store_cache" default:"false"`
}

func main() {
	var cfg config

	if err := envconfig.Process("cli", &cfg); err != nil {
		log.Error((err.Error()))
	}

	client, err := mgo.Connect(context.TODO(), options.Client().ApplyURI(cfg.MongoUri))
	if err != nil {
		log.Error(err)
	}

	var cache storecache.Cache

	if cfg.StoreCache {
		cache = storecache.NewRedisCache()
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
			username, err := svc.UserCreate(Parameters{
				Username: args[0],
				Password: args[1],
				Email:    args[2],
			})
			if err != nil {
				return err
			}

			fmt.Println("User added: ", username)

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

				ns, err := svc.NamespaceCreate(Parameters{
					Namespace: args[0],
					Username:  args[1],
					TenantID:  args[2],
				})
				if err != nil {
					return err
				}

				fmt.Println("Namespace added: ", ns.Name)
				fmt.Println("Owner: ", ns.Owner)
				fmt.Println("Tenant ID: ", ns.TenantID)

				return nil
			},
		},

		&cobra.Command{
			Use:   "add-user-namespace",
			Short: "Usage: <username> <namespace>",
			Args:  cobra.ExactArgs(2),
			RunE: func(cmd *cobra.Command, args []string) error {
				ns, err := svc.NamespaceAddMember(Parameters{
					Username:  args[0],
					Namespace: args[1],
				})
				if err != nil {
					return err
				}

				fmt.Println("User: ", ns.Owner)
				fmt.Println("added to namespace: ", ns.Name)

				return nil
			},
		},

		&cobra.Command{
			Use:   "del-namespace",
			Short: "Usage: <namespace>",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				if err := svc.NamespaceDelete(Parameters{
					Namespace: args[0],
				}); err != nil {
					return err
				}

				fmt.Println("Namespace deleted")

				return nil
			},
		},

		&cobra.Command{
			Use:   "del-user",
			Short: "Usage: <username>",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				if err := svc.UserDelete(Parameters{
					Username: args[0],
				}); err != nil {
					return err
				}

				fmt.Println("User deleted")

				return nil
			},
		},

		&cobra.Command{
			Use:   "del-user-namespace",
			Short: "Usage <username> <namespace>",
			Args:  cobra.ExactArgs(2),
			RunE: func(cmd *cobra.Command, args []string) error {
				ns, err := svc.NamespaceRemoveMember(Parameters{
					Username:  args[0],
					Namespace: args[1],
				})
				if err != nil {
					return err
				}

				fmt.Println("User: ", ns.Owner)
				fmt.Println("removed from namespace: ", ns.Name)

				return nil
			},
		},

		&cobra.Command{
			Use:   "reset-user-password",
			Short: "Usage: <username> <password>",
			Args:  cobra.ExactArgs(2),
			RunE: func(cmd *cobra.Command, args []string) error {
				if err := svc.UserUpdate(Parameters{
					Username: args[0],
					Password: args[1],
				}); err != nil {
					return err
				}

				fmt.Println("Password changed")

				return nil
			},
		})

	if err := rootCmd.Execute(); err != nil {
		log.Error(err)
	}
}
