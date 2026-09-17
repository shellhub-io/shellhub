package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"

	"github.com/shellhub-io/shellhub/pkg/uuid"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/spf13/cobra"
)

const (
	adminUsername  = "e2e-admin"
	adminPassword  = "e2e-password"
	adminEmail     = "admin@e2e.test"
	adminNamespace = "e2e"
)

type stateFile struct {
	Name    string            `json:"name"`
	Edition string            `json:"edition"`
	Files   []string          `json:"files"`
	Envs    map[string]string `json:"envs"`
}

func statePath(name string) string {
	return filepath.Join(".stack", name+".json")
}

func logPath(name string) string {
	return filepath.Join(".stack", name+".log")
}

func readState(name string) (*stateFile, error) {
	data, err := os.ReadFile(statePath(name))
	if err != nil {
		return nil, err
	}

	var s stateFile
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}

	return &s, nil
}

func writeState(s *stateFile, name string) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(statePath(name), data, 0o600)
}

func printExports(w io.Writer, s *stateFile) error {
	baseURL := "http://localhost:" + s.Envs["SHELLHUB_HTTP_PORT"]

	for _, line := range []string{
		"E2E_BASE_URL=" + baseURL,
		"E2E_EDITION=" + s.Edition,
		"E2E_ADMIN_USER=" + adminUsername,
		"E2E_ADMIN_PASSWORD=" + adminPassword,
		"E2E_ADMIN_NAMESPACE=" + adminNamespace,
	} {
		if _, err := fmt.Fprintf(w, "export %s\n", line); err != nil {
			return err
		}
	}

	return nil
}

func saveLogs(ctx context.Context, cmd *cobra.Command, stack *environment.Stack, name string) {
	lp := logPath(name)
	cmd.PrintErrf("Saving logs to %s...\n", lp)

	f, err := os.Create(lp) //nolint:gosec // path derived from validated --name flag
	if err != nil {
		cmd.PrintErrf("Could not create log file: %v\n", err)

		return
	}
	defer func() { _ = f.Close() }()

	if err := stack.Logs(ctx, f); err != nil {
		cmd.PrintErrf("Could not save logs: %v\n", err)
	}
}

func teardown(ctx context.Context, cmd *cobra.Command, name string) error {
	existing, err := readState(name)
	if err != nil {
		cmd.PrintErrf("No state file for %q, cleaning up orphans...\n", name)

		orphan, err := environment.Attach(ctx, "shellhub-e2e-"+name, []string{"../docker-compose.yml"}, nil)
		if err != nil {
			cmd.PrintErrf("Could not attach to orphans: %v\n", err)

			return nil
		}

		saveLogs(ctx, cmd, orphan, name)

		if err := orphan.Down(ctx); err != nil {
			cmd.PrintErrf("Could not clean up orphans: %v\n", err)
		}

		return nil
	}

	stack, err := environment.Attach(ctx, existing.Name, existing.Files, existing.Envs)
	if err != nil {
		return fmt.Errorf("attaching to stack: %w", err)
	}

	saveLogs(ctx, cmd, stack, name)

	cmd.PrintErrf("Tearing down %q...\n", existing.Name)

	if err := stack.Down(ctx); err != nil {
		return fmt.Errorf("tearing down stack: %w", err)
	}

	if err := os.Remove(statePath(name)); err != nil {
		cmd.PrintErrf("Could not remove state file: %v\n", err)
	}

	return nil
}

func main() {
	_ = os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")

	var (
		name     string
		cloudDir string
	)

	validName := regexp.MustCompile(`^[a-z0-9][a-z0-9_-]*$`)

	root := &cobra.Command{
		Use:           "stack",
		Short:         "Manage a ShellHub e2e stack",
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
			if !validName.MatchString(name) {
				return fmt.Errorf("invalid stack name %q: must match %s", name, validName.String())
			}

			return nil
		},
	}

	root.PersistentFlags().StringVar(&name, "name", "default", "stack name")
	root.PersistentFlags().StringVar(&cloudDir, "cloud-dir", "../../cloud", "path to the cloud repo")

	var (
		editionFlag string
		httpPort    string
	)

	upCmd := &cobra.Command{
		Use:   "up",
		Short: "Bring the e2e stack up",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			stdout := cmd.OutOrStdout()

			edition, err := environment.ParseEdition(editionFlag)
			if err != nil {
				return err
			}

			if err := os.MkdirAll(".stack", 0o750); err != nil {
				return err
			}

			cmd.PrintErrln("Bundling OpenAPI schema...")

			if err := environment.BundleOpenAPI(ctx, edition, ".."); err != nil {
				return fmt.Errorf("bundling OpenAPI schema: %w", err)
			}

			projectName := "shellhub-e2e-" + name

			if err := teardown(ctx, cmd, name); err != nil {
				return err
			}

			cmd.PrintErrln("Generating keys...")

			if err := environment.GenerateKeys(".."); err != nil {
				return fmt.Errorf("generating keys: %w", err)
			}

			cfg := environment.Config{
				Edition:  edition,
				Name:     projectName,
				HTTPPort: httpPort,
				CloudDir: cloudDir,
			}

			cmd.PrintErrf("Starting stack %q (edition=%s)...\n", projectName, edition)

			stack, err := environment.Up(ctx, cfg)
			if err != nil {
				return fmt.Errorf("bringing stack up: %w", err)
			}

			cmd.PrintErrln("Waiting for API...")

			if err := stack.AwaitAPI(ctx); err != nil {
				return fmt.Errorf("waiting for API: %w", err)
			}

			cmd.PrintErrln("Seeding admin user...")

			if err := stack.NewUser(ctx, adminUsername, adminEmail, adminPassword); err != nil {
				return fmt.Errorf("creating admin user: %w", err)
			}

			if err := stack.NewNamespace(ctx, adminUsername, adminNamespace, uuid.Generate(), ""); err != nil {
				return fmt.Errorf("creating admin namespace: %w", err)
			}

			cmd.PrintErrln("Verifying login...")

			if _, err := stack.AwaitLogin(ctx, adminUsername, adminPassword); err != nil {
				return fmt.Errorf("verifying admin login: %w", err)
			}

			state := &stateFile{
				Name:    projectName,
				Edition: string(edition),
				Files:   stack.Files(),
				Envs:    stack.Envs(),
			}

			if err := writeState(state, name); err != nil {
				return fmt.Errorf("writing state file: %w", err)
			}

			cmd.PrintErrln("Stack ready.")

			return printExports(stdout, state)
		},
	}

	upCmd.Flags().StringVar(&editionFlag, "edition", "community", "ShellHub edition (community, enterprise, cloud)")
	upCmd.Flags().StringVar(&httpPort, "http-port", "", "host HTTP port (random if empty)")

	downCmd := &cobra.Command{
		Use:   "down",
		Short: "Tear the e2e stack down",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := teardown(cmd.Context(), cmd, name); err != nil {
				return err
			}

			cmd.PrintErrln("Done.")

			return nil
		},
	}

	root.AddCommand(upCmd, downCmd)

	if err := root.ExecuteContext(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
