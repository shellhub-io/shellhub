package cmd

import (
	"github.com/shellhub-io/shellhub/cli/services"
	"github.com/spf13/cobra"
)

// DeprecatedCommands writes a list of deprecated commands to a parend cmd.
// These commands will be romved in a future release.
func DeprecatedCommands(cmd *cobra.Command, service services.Services) {
	cmd.AddCommand(&cobra.Command{
		Deprecated: "This command is deprecated and will be removed in a future release.",
		Use:        "add-user",
		Short:      "Usage: <username> <password> <email>",
		Args:       cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			user, err := service.UserCreate(cmd.Context(), args[0], args[1], args[2])
			if err != nil {
				return err
			}
			cmd.Println("User added!")
			cmd.Println("name:", user.Name)
			cmd.Println("username:", user.Username)
			cmd.Println("email:", user.Email)

			return nil
		},
	},
		&cobra.Command{
			Deprecated: "This command is deprecated and will be removed in a future release.",
			Use:        "del-user",
			Short:      "Usage: <username>",
			Args:       cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				if err := service.UserDelete(cmd.Context(), args[0]); err != nil {
					return err
				}

				cmd.Println("User deleted")

				return nil
			},
		},
		&cobra.Command{
			Deprecated: "This command is deprecated and will be removed in a future release.",
			Use:        "reset-user-password",
			Short:      "Usage: <username> <password>",
			Args:       cobra.ExactArgs(2),
			RunE: func(cmd *cobra.Command, args []string) error {
				if err := service.UserUpdate(cmd.Context(), args[0], args[1]); err != nil {
					return err
				}

				cmd.Println("Password changed")

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

				namespace, err := service.NamespaceCreate(cmd.Context(), args[0], args[1], args[2])
				if err != nil {
					return err
				}

				cmd.Println("Namespace added:", namespace.Name)
				cmd.Println("Owner:", namespace.Owner)
				cmd.Println("Tenant ID:", namespace.TenantID)

				return nil
			},
		},
		&cobra.Command{
			Deprecated: "This command is deprecated and will be removed in a future release.",
			Use:        "add-user-namespace",
			Short:      "Usage: <username> <namespace> <role>",
			Args:       cobra.ExactArgs(3),
			RunE: func(cmd *cobra.Command, args []string) error {
				ns, err := service.NamespaceAddMember(cmd.Context(), args[0], args[1], args[2])
				if err != nil {
					return err
				}

				cmd.Println("User:", ns.Owner)
				cmd.Println("added to namespace:", ns.Name)
				cmd.Println("with role:", args[2])

				return nil
			},
		},
		&cobra.Command{
			Deprecated: "This command is deprecated and will be removed in a future release.",
			Use:        "del-user-namespace",
			Short:      "Usage <username> <namespace>",
			Args:       cobra.ExactArgs(2),
			RunE: func(cmd *cobra.Command, args []string) error {
				ns, err := service.NamespaceRemoveMember(cmd.Context(), args[0], args[1])
				if err != nil {
					return err
				}

				cmd.Println("User:", ns.Owner)
				cmd.Println("removed from namespace:", ns.Name)

				return nil
			},
		},
		&cobra.Command{
			Deprecated: "This command is deprecated and will be removed in a future release.",
			Use:        "del-namespace",
			Short:      "Usage: <namespace>",
			Args:       cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				if err := service.NamespaceDelete(cmd.Context(), args[0]); err != nil {
					return err
				}

				cmd.Println("Namespace deleted")

				return nil
			},
		})
}
