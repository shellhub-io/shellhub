package main

import (
	"context"

	"github.com/kelseyhightower/envconfig"
	storecache "github.com/shellhub-io/shellhub/api/cache"
	"github.com/shellhub-io/shellhub/api/store/mongo"
	"github.com/shellhub-io/shellhub/pkg/loglevel"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	mgo "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
)

type config struct {
	MongoURI string `envconfig:"mongo_uri" default:"mongodb://mongo:27017/main"`
	RedisURI string `envconfig:"redis_uri" default:"redis://redis:6379"`
}

func init() {
	loglevel.SetLogLevel()
}

func main() {
	var cfg config
	if err := envconfig.Process("cli", &cfg); err != nil {
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

	cache, err := storecache.NewRedisCache(cfg.RedisURI)
	if err != nil {
		log.Fatal(err)
	}

	services := NewService(mongo.NewStore(client.Database(connStr.Database), cache))

	rootCmd := &cobra.Command{Use: "cli"}

	userCmd := &cobra.Command{
		Use:   "user",
		Short: "Manage users",
		Long:  `Manage users`,
	}
	userCmd.AddCommand(&cobra.Command{
		Use:     "create <username> <password> <email>",
		Short:   "Create an user",
		Long:    `Create an user`,
		Example: `cli user create shellhub password`,
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			user, err := services.UserCreate(args[0], args[1], args[2])
			if err != nil {
				return err
			}

			cmd.Println("User created successfully")
			cmd.Println("Username:", user.Username)
			cmd.Println("Email:", user.Email)

			return nil
		},
	})
	userCmd.AddCommand(&cobra.Command{
		Use:     "delete <username>",
		Short:   "Delete an user",
		Long:    `Delete an user`,
		Example: `cli user delete shellhub`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := services.UserDelete(args[0]); err != nil {
				return err
			}

			cmd.Println("User deleted successfully")
			cmd.Println("Username:", args[0])

			return nil
		},
	})

	userCmd.AddCommand(&cobra.Command{
		Use:     "password <username> <password>",
		Short:   "Change user password",
		Long:    `Change user password`,
		Example: `cli user password shellhub password`,
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := services.UserUpdate(args[0], args[1]); err != nil {
				return err
			}

			cmd.Println("User password changed successfully")
			cmd.Println("Username:", args[0])

			return nil
		},
	})

	namespaceCmd := &cobra.Command{
		Use:   "namespace",
		Short: "Manage namespaces",
		Long:  `Manage namespaces`,
	}
	namespaceCmd.AddCommand(&cobra.Command{
		Use:     "create <namespace> <owner>",
		Short:   "create a namespace",
		Long:    `create a namespace`,
		Example: `cli namespace create shellhubspace shellhub`,
		Args:    cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Avoid panic when TenantID isn't provided.
			if len(args) == 2 {
				args = append(args, "")
			}

			namespace, err := services.NamespaceCreate(args[0], args[1], args[2])
			if err != nil {
				return err
			}

			cmd.Println("Namespace created successfully")
			cmd.Println("Namespace:", namespace.Name)
			cmd.Println("Tenant:", namespace.TenantID)
			cmd.Println("Owner:", namespace.Owner)

			return nil
		},
	})
	namespaceCmd.AddCommand(&cobra.Command{
		Use:     "delete <namespace>",
		Short:   "Delete a namespace",
		Long:    `Delete a namespace`,
		Example: `cli namespace delete shellhubspace`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			namespace, err := services.NamespaceRemoveMember(args[0], args[1])
			if err != nil {
				return err
			}

			cmd.Println("Namespace deleted successfully")
			cmd.Println("Namespace:", namespace.Name)
			cmd.Println("Tenant:", namespace.TenantID)
			cmd.Println("Owner:", namespace.Owner)

			return nil
		},
	})

	memberCmd := &cobra.Command{
		Use:   "member",
		Short: "Manage members",
		Long:  `Manage members`,
	}
	memberCmd.AddCommand(&cobra.Command{
		Use:     "add <username> <namespace> <role>",
		Short:   "Add a member",
		Long:    `Add a member`,
		Example: `cli member add shellhub shellhubspace`,
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			namespace, err := services.NamespaceAddMember(args[0], args[1], args[2])
			if err != nil {
				return err
			}

			cmd.Println("Member added successfully")
			cmd.Println("Namespace:", namespace.Name)
			cmd.Println("Tenant:", namespace.TenantID)
			cmd.Println("Member:", args[0])
			cmd.Println("Role:", args[2])

			return nil
		},
	})
	memberCmd.AddCommand(&cobra.Command{
		Use:     "remove <username> <namespace>",
		Short:   "Remove a member",
		Long:    `Remove a member`,
		Example: `cli member remove shellhub shellhubspace`,
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			namespace, err := services.NamespaceRemoveMember(args[0], args[1])
			if err != nil {
				return err
			}

			cmd.Println("Member removed successfully")
			cmd.Println("Namespace:", namespace.Name)
			cmd.Println("Tenant:", namespace.TenantID)
			cmd.Println("Member:", args[0])

			return nil
		},
	})

	rootCmd.AddCommand(userCmd)
	rootCmd.AddCommand(namespaceCmd)
	rootCmd.AddCommand(memberCmd)

	rootCmd.AddCommand(&cobra.Command{
		Deprecated: "This command is deprecated and will be removed in a future release.",
		Use:        "add-user",
		Short:      "Usage: <username> <password> <email>",
		Args:       cobra.ExactArgs(3),
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
			Deprecated: "This command is deprecated and will be removed in a future release.",
			Use:        "del-user",
			Short:      "Usage: <username>",
			Args:       cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				if err := services.UserDelete(args[0]); err != nil {
					return err
				}

				rootCmd.Println("User deleted")

				return nil
			},
		},
		&cobra.Command{
			Deprecated: "This command is deprecated and will be removed in a future release.",
			Use:        "reset-user-password",
			Short:      "Usage: <username> <password>",
			Args:       cobra.ExactArgs(2),
			RunE: func(cmd *cobra.Command, args []string) error {
				if err := services.UserUpdate(args[0], args[1]); err != nil {
					return err
				}

				rootCmd.Println("Password changed")

				return nil
			},
		},
		&cobra.Command{
			Deprecated: "This command is deprecated and will be removed in a future release.",
			Use:        "add-namespace",
			Short:      "Usage: <namespace> <owner>",
			Args:       cobra.RangeArgs(2, 3),
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
			Deprecated: "This command is deprecated and will be removed in a future release.",
			Use:        "add-user-namespace",
			Short:      "Usage: <username> <namespace> <role>",
			Args:       cobra.ExactArgs(3),
			RunE: func(cmd *cobra.Command, args []string) error {
				ns, err := services.NamespaceAddMember(args[0], args[1], args[2])
				if err != nil {
					return err
				}

				rootCmd.Println("User:", ns.Owner)
				rootCmd.Println("added to namespace:", ns.Name)
				rootCmd.Println("with role:", args[2])

				return nil
			},
		},
		&cobra.Command{
			Deprecated: "This command is deprecated and will be removed in a future release.",
			Use:        "del-user-namespace",
			Short:      "Usage <username> <namespace>",
			Args:       cobra.ExactArgs(2),
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
			Deprecated: "This command is deprecated and will be removed in a future release.",
			Use:        "del-namespace",
			Short:      "Usage: <namespace>",
			Args:       cobra.ExactArgs(1),
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
