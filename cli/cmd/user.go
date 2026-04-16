package cmd

import (
	"fmt"
	"text/tabwriter"

	"github.com/shellhub-io/shellhub/cli/pkg/inputs"
	"github.com/shellhub-io/shellhub/cli/services"
	"github.com/spf13/cobra"
)

// UserCommands creates and returns a Cobra command for user management.
// It registers user-related subcommands and uses the provided service
// to handle the underlying business logic.
func UserCommands(service services.Services) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user",
		Short: "Manage user accounts",
		Long:  `Manage user accounts in the system, including creating, listing, deleting, and other user-related operations.`,
	}

	cmd.AddCommand(
		userCreate(service),
		userResetPassword(service),
		userDelete(service),
		userList(service),
	)

	return cmd
}

func userCreate(service services.Services) *cobra.Command {
	var admin bool

	cmd := &cobra.Command{
		Use:   "create <username> <password> <email>",
		Args:  cobra.ExactArgs(3),
		Short: "Create a user",
		Long: `Create a new user with the specified username, password, and email.
The username must be unique, and the password must meet the system's security requirements.`,
		Example: `cli user create john_doe Secret123!- john.doe@test.com
cli user create john_doe Secret123!- john.doe@test.com --admin`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var input inputs.UserCreate

			if err := bind(args, &input); err != nil {
				return err
			}

			input.Admin = admin

			user, err := service.UserCreate(cmd.Context(), &input)
			if err != nil {
				return err
			}

			cmd.Println("User created successfully")
			cmd.Println("Username:", user.Username)
			cmd.Println("Email:", user.Email)

			return nil
		},
	}

	cmd.Flags().BoolVar(&admin, "admin", false, "Create user with admin privileges")

	return cmd
}

func userResetPassword(service services.Services) *cobra.Command {
	return &cobra.Command{
		Use:     "password <username> <password>",
		Args:    cobra.ExactArgs(2),
		Short:   "Change a user's password",
		Long:    `Updates the password for an existing user identified by the given username.`,
		Example: `cli user password john_doe Secret123!-`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var input inputs.UserUpdate

			if err := bind(args, &input); err != nil {
				return err
			}

			if err := service.UserUpdate(cmd.Context(), &input); err != nil {
				return err
			}

			cmd.Println("User password changed successfully")
			cmd.Println("Username:", input.Username)

			return nil
		},
	}
}

func userDelete(service services.Services) *cobra.Command {
	return &cobra.Command{
		Use:     "delete <username>",
		Args:    cobra.ExactArgs(1),
		Short:   "Delete a user",
		Long:    `Deletes a user and all associated data from the system based on the provided username.`,
		Example: `cli user delete john_doe`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var input inputs.UserDelete

			if err := bind(args, &input); err != nil {
				return err
			}

			if err := service.UserDelete(cmd.Context(), &input); err != nil {
				return err
			}

			cmd.Println("User deleted successfully")
			cmd.Println("Username:", input.Username)

			return nil
		},
	}
}

func userList(service services.Services) *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List all users",
		Long:    "List all users in the system",
		Example: `cli user list
cli user ls | head -n 5
cli user ls | tail -n 5
cli user ls | grep admin`,
		RunE: func(cmd *cobra.Command, args []string) error {
			users, err := service.UserList(cmd.Context())
			if err != nil {
				return err
			}

			if len(users) == 0 {
				cmd.Println("No users found. Use `cli user create --help` for more information.")

				return nil
			}

			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "USERNAME\tEMAIL\tROLE")
			for _, u := range users {
				role := "user"
				if u.Admin {
					role = "admin"
				}
				fmt.Fprintf(w, "%s\t%s\t%s\n", u.Username, u.Email, role)
			}
			w.Flush()

			return nil
		},
	}
}
