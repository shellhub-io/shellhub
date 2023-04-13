package main

import (
	"context"
	"errors"
	"reflect"

	"github.com/kelseyhightower/envconfig"
	"github.com/shellhub-io/shellhub/api/store/mongo"
	"github.com/shellhub-io/shellhub/cli/services"
	storecache "github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/loglevel"
	"github.com/shellhub-io/shellhub/pkg/validator"
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

func bind(args []string, input interface{}) error {
	typeOf := reflect.TypeOf(input)
	valueOf := reflect.ValueOf(input)

	if typeOf.Kind() != reflect.Ptr || typeOf.Elem().Kind() != reflect.Struct {
		return errors.New("input must be a pointer to a structure")
	}

	for i := 0; i < len(args); i++ {
		valueOf.Elem().
			FieldByName(typeOf.Elem().FieldByIndex([]int{i}).Name).
			SetString(args[i])
	}

	return nil
}

func validate(input interface{}) error {
	v := validator.New()
	if ok, err := v.Struct(input); !ok || err != nil {
		return validator.GetFirstFieldError(errors.Unwrap(err))
	}

	return nil
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

	service := services.NewService(mongo.NewStore(client.Database(connStr.Database), cache))

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
		Example: `cli user create myuser mypassword`,
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			var input struct {
				Username string `validate:"required,username"`
				Password string `validate:"required,password"`
				Email    string `validate:"required,email"`
			}

			if err := bind(args, &input); err != nil {
				return err
			}

			if err := validate(input); err != nil {
				return err
			}

			user, err := service.UserCreate(input.Username, input.Password, input.Email)
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
		Example: `cli user delete myuser`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var input struct {
				Username string `validate:"required,username"`
			}

			if err := bind(args, &input); err != nil {
				return err
			}

			if err := validate(input); err != nil {
				return err
			}

			if err := service.UserDelete(input.Username); err != nil {
				return err
			}

			cmd.Println("User deleted successfully")
			cmd.Println("Username:", input.Username)

			return nil
		},
	})
	userCmd.AddCommand(&cobra.Command{
		Use:     "password <username> <password>",
		Short:   "Change user password",
		Long:    `Change user password`,
		Example: `cli user password myuser mynewpassword`,
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			var input struct {
				Username string `validate:"required,username"`
				Password string `validate:"required,password"`
			}

			if err := bind(args, &input); err != nil {
				return err
			}

			if err := validate(input); err != nil {
				return err
			}

			if err := service.UserUpdate(input.Username, input.Password); err != nil {
				return err
			}

			cmd.Println("User password changed successfully")
			cmd.Println("Username:", input.Username)

			return nil
		},
	})

	namespaceCmd := &cobra.Command{
		Use:   "namespace",
		Short: "Manage namespaces",
		Long:  `Manage namespaces`,
	}
	namespaceCmd.AddCommand(&cobra.Command{
		Use:     "create <namespace> <owner> [tenant]",
		Short:   "create a namespace",
		Long:    `create a namespace`,
		Example: `cli namespace create mynamespace myuser`,
		Args:    cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Avoid panic when TenantID isn't provided.
			if len(args) == 2 {
				args = append(args, "")
			}

			var input struct {
				Namespace string
				Owner     string `validate:"required,username"`
				TenantID  string `validate:"required,uuid"`
			}

			if err := bind(args, &input); err != nil {
				return err
			}

			if err := validate(input); err != nil {
				return err
			}

			namespace, err := service.NamespaceCreate(input.Namespace, input.Owner, input.TenantID)
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
		Example: `cli namespace delete mynamespace`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var input struct {
				Namespace string
			}

			if err := bind(args, &input); err != nil {
				return err
			}

			if err := validate(input); err != nil {
				return err
			}

			if err := service.NamespaceDelete(input.Namespace); err != nil {
				return err
			}

			cmd.Println("Namespace deleted successfully")
			cmd.Println("Namespace:", input.Namespace)

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
		Example: `cli member add myuser mynamespace observer`,
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			var input struct {
				Username  string `validate:"required,username"`
				Namespace string
				Role      string
			}

			if err := bind(args, &input); err != nil {
				return err
			}

			if err := validate(input); err != nil {
				return err
			}

			namespace, err := service.NamespaceAddMember(input.Username, input.Namespace, input.Role)
			if err != nil {
				return err
			}

			cmd.Println("Member added successfully")
			cmd.Println("Namespace:", namespace.Name)
			cmd.Println("Tenant:", namespace.TenantID)
			cmd.Println("Member:", input.Username)
			cmd.Println("Role:", input.Role)

			return nil
		},
	})
	memberCmd.AddCommand(&cobra.Command{
		Use:     "remove <username> <namespace>",
		Short:   "Remove a member",
		Long:    `Remove a member`,
		Example: `cli member remove memberusername mynamespace`,
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			var input struct {
				Username  string `validate:"required,username"`
				Namespace string
			}

			if err := bind(args, &input); err != nil {
				return err
			}

			if err := validate(input); err != nil {
				return err
			}

			namespace, err := service.NamespaceRemoveMember(input.Username, input.Namespace)
			if err != nil {
				return err
			}

			cmd.Println("Member removed successfully")
			cmd.Println("Namespace:", namespace.Name)
			cmd.Println("Tenant:", namespace.TenantID)
			cmd.Println("Member:", input.Username)

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
			user, err := service.UserCreate(args[0], args[1], args[2])
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
				if err := service.UserDelete(args[0]); err != nil {
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
				if err := service.UserUpdate(args[0], args[1]); err != nil {
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

				namespace, err := service.NamespaceCreate(args[0], args[1], args[2])
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
				ns, err := service.NamespaceAddMember(args[0], args[1], args[2])
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
				ns, err := service.NamespaceRemoveMember(args[0], args[1])
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
				if err := service.NamespaceDelete(args[0]); err != nil {
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
