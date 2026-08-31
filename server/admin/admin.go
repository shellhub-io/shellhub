// Package admin provides the "admin" command group: direct, database-level
// management of users and namespaces, bypassing the API.
//
// It is registered on the server's root command by the app package, and reached
// in practice through the bin/cli wrapper, which injects the "admin" token so
// `./bin/cli user create ...` keeps working.
package admin

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/server/admin/services"
	"github.com/shellhub-io/shellhub/server/api/store/pg"
	pgoptions "github.com/shellhub-io/shellhub/server/api/store/pg/options"
	"github.com/spf13/cobra"
)

type env struct {
	PostgresHost       string `env:"POSTGRES_HOST,default=postgres"`
	PostgresPort       string `env:"POSTGRES_PORT,default=5432"`
	PostgresUsername   string `env:"POSTGRES_USERNAME,default=admin"`
	PostgresPassword   string `env:"POSTGRES_PASSWORD,default=admin"`
	PostgresDatabase   string `env:"POSTGRES_DATABASE,default=main"`
	PostgresLogLevel   string `env:"POSTGRES_LOG_LEVEL,default=INFO"`
	PostgresLogVerbose bool   `env:"POSTGRES_LOG_VERBOSE,default=false"`
}

type serviceFunc func() services.Services

// Command returns the "admin" command group.
func Command() *cobra.Command {
	var service services.Services

	cmd := &cobra.Command{
		Use:   "admin",
		Short: "Manage users and namespaces directly in the database",
		Long: `Manage users and namespaces by writing to the database directly, without going
through the API. Intended for operators administering a ShellHub instance.`,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			s, err := connect(cmd.Context())
			if err != nil {
				return err
			}

			service = s

			return nil
		},
	}

	get := serviceFunc(func() services.Services { return service })

	cmd.AddCommand(
		userCommands(get),
		namespaceCommands(get),
	)

	return cmd
}

func connect(ctx context.Context) (services.Services, error) {
	cfg, err := envs.ParseWithPrefix[env]("ADMIN_")
	if err != nil {
		return nil, err
	}

	uri := pg.URI(
		cfg.PostgresHost,
		cfg.PostgresPort,
		cfg.PostgresUsername,
		cfg.PostgresPassword,
		cfg.PostgresDatabase,
	)

	store, err := pg.New(ctx, uri, pgoptions.Log(cfg.PostgresLogLevel, cfg.PostgresLogVerbose))
	if err != nil {
		return nil, err
	}

	return services.NewService(store), nil
}
