package admin

import (
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/shellhub-io/shellhub/server/admin/inputs"
	"github.com/spf13/cobra"
)

func userCommands(service serviceFunc) *cobra.Command {
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

func userCreate(service serviceFunc) *cobra.Command {
	var admin bool

	cmd := &cobra.Command{
		Use:   "create <username> <password> <email>",
		Args:  cobra.ExactArgs(3),
		Short: "Create a user",
		Long: `Create a new user with the specified username, password, and email.
The username must be unique, and the password must meet the system's security requirements.`,
		Example: `./bin/cli user create john_doe Secret123!- john.doe@test.com
./bin/cli user create john_doe Secret123!- john.doe@test.com --admin`,
		RunE: func(cmd *cobra.Command, args []string) error {
			input := inputs.UserCreate{
				Username: strings.ToLower(args[0]),
				Password: args[1],
				Email:    strings.ToLower(args[2]),
				Admin:    admin,
			}

			if err := validateInput(input); err != nil {
				return err
			}

			user, err := service().UserCreate(cmd.Context(), &input)
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

func userResetPassword(service serviceFunc) *cobra.Command {
	return &cobra.Command{
		Use:     "password <username> <password>",
		Args:    cobra.ExactArgs(2),
		Short:   "Change a user's password",
		Long:    `Updates the password for an existing user identified by the given username.`,
		Example: `./bin/cli user password john_doe Secret123!-`,
		RunE: func(cmd *cobra.Command, args []string) error {
			input := inputs.UserUpdate{
				Username: strings.ToLower(args[0]),
				Password: args[1],
			}

			if err := validateInput(input); err != nil {
				return err
			}

			if err := service().UserUpdate(cmd.Context(), &input); err != nil {
				return err
			}

			cmd.Println("User password changed successfully")
			cmd.Println("Username:", input.Username)

			return nil
		},
	}
}

func userDelete(service serviceFunc) *cobra.Command {
	return &cobra.Command{
		Use:     "delete <username>",
		Args:    cobra.ExactArgs(1),
		Short:   "Delete a user",
		Long:    `Deletes a user and all associated data from the system based on the provided username.`,
		Example: `./bin/cli user delete john_doe`,
		RunE: func(cmd *cobra.Command, args []string) error {
			input := inputs.UserDelete{
				Username: strings.ToLower(args[0]),
			}

			if err := validateInput(input); err != nil {
				return err
			}

			if err := service().UserDelete(cmd.Context(), &input); err != nil {
				return err
			}

			cmd.Println("User deleted successfully")
			cmd.Println("Username:", input.Username)

			return nil
		},
	}
}

func userList(service serviceFunc) *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List all users",
		Long:    "List all users in the system",
		Example: `./bin/cli user list
./bin/cli user ls | head -n 5
./bin/cli user ls | tail -n 5
./bin/cli user ls | grep admin`,
		RunE: func(cmd *cobra.Command, args []string) error {
			users, err := service().UserList(cmd.Context())
			if err != nil {
				return err
			}

			if len(users) == 0 {
				cmd.Println("No users found. Use `./bin/cli user create --help` for more information.")

				return nil
			}

			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			_, _ = fmt.Fprintln(w, "USERNAME\tEMAIL\tROLE")
			for _, u := range users {
				role := "user"
				if u.Admin {
					role = "admin"
				}
				_, _ = fmt.Fprintf(w, "%s\t%s\t%s\n", u.Username, u.Email, role)
			}
			_ = w.Flush()

			return nil
		},
	}
}
