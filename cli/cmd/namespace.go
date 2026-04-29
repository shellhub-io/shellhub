package cmd

import (
	"fmt"
	"text/tabwriter"

	"github.com/shellhub-io/shellhub/cli/pkg/inputs"
	"github.com/shellhub-io/shellhub/cli/services"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	"github.com/spf13/cobra"
)

// NamespaceCommands creates and returns a Cobra command for namespace management.
// It registers namespace-related subcommands and uses the provided service
// to handle the underlying business logic.
func NamespaceCommands(service services.Services) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "namespace",
		Short: "Manage namespaces",
		Long:  `Provides an interface for managing namespaces within the system, such as creating new namespaces or deleting existing ones.`,
	}

	cmd.AddCommand(
		namespaceCreate(service),
		namespaceDelete(service),
		namespaceList(service),
		namespaceInspect(service),
		memberCommands(service),
	)

	return cmd
}

func namespaceCreate(service services.Services) *cobra.Command {
	cmdNamespace := &cobra.Command{
		Use:   "create <namespace> <owner> [tenant]",
		Short: "Create a namespace",
		Long: `Creates a new namespace in the system using the provided namespace name, associated owner's username, and an optional tenant ID and Type.
The owner must be a valid username within the system. If a tenant ID is provided, it should be in UUID format.`,
		Example: `cli namespace create dev john_doe --type=team`,
		Args:    cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			namespaceType, err := cmd.Flags().GetString("type")
			if err != nil {
				return err
			}

			input := inputs.NamespaceCreate{
				Namespace: args[0],
				Owner:     args[1],
				TenantID:  "",
				Type:      namespaceType,
			}

			if len(args) == 3 {
				input.TenantID = args[2]
			}

			namespace, err := service.NamespaceCreate(cmd.Context(), &input)
			if err != nil {
				return err
			}

			cmd.Println("Namespace created successfully")
			cmd.Println("Namespace:", namespace.Name)
			cmd.Println("Tenant:", namespace.TenantID)
			cmd.Println("Owner:", namespace.Owner)
			cmd.Println("Type:", namespace.Type)

			return nil
		},
	}

	cmdNamespace.PersistentFlags().String("type", "team", "type")

	return cmdNamespace
}

func namespaceDelete(service services.Services) *cobra.Command {
	return &cobra.Command{
		Use:     "delete <namespace>",
		Short:   "Delete a namespace",
		Long:    `Deletes a namespace and all associated data from the system based on the provided name.`,
		Example: `cli namespace delete dev`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			input := inputs.NamespaceDelete{
				Namespace: args[0],
			}

			if err := service.NamespaceDelete(cmd.Context(), &input); err != nil {
				return err
			}

			cmd.Println("Namespace deleted successfully")
			cmd.Println("Namespace:", input.Namespace)

			return nil
		},
	}
}

func namespaceList(service services.Services) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List namespaces",
		Long:    "List all namespaces in the system",
		Example: `cli namespace list
cli namespace ls
cli namespace ls -q
cli namespace ls -q tenant-id`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			validFields := map[string]bool{
				"name":      true,
				"tenant-id": true,
			}

			quiet, err := cmd.Flags().GetBool("quiet")
			if err != nil {
				return err
			}

			field := "name"
			if len(args) == 1 {
				field = args[0]
			}

			if !quiet && len(args) == 1 {
				return fmt.Errorf("field argument requires -q")
			}

			if !validFields[field] {
				return fmt.Errorf("invalid field: %s (allowed: name, tenant-id)", field)
			}

			namespaces, err := service.NamespaceList(cmd.Context())
			if err != nil {
				return err
			}

			if len(namespaces) == 0 {
				if !quiet {
					fmt.Fprintln(out, "No namespaces to list")
				}

				return nil
			}

			if quiet {
				for _, ns := range namespaces {
					var v string
					if field == "tenant-id" {
						v = ns.TenantID
					} else {
						v = ns.Name
					}
					fmt.Fprintln(out, v)
				}

				return nil
			}

			// non-quiet output
			w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "NAME\tTENANT ID\tTYPE")

			for _, ns := range namespaces {
				fmt.Fprintf(w, "%s\t%s\t%s\n", ns.Name, ns.TenantID, ns.Type)
			}
			w.Flush()

			return nil
		},
	}

	cmd.Flags().BoolP("quiet", "q", false,
		"Output only a single field (default: name, options: name, tenant-id)")

	return cmd
}

func namespaceInspect(service services.Services) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "inspect <namespace>",
		Short: "Inspect a namespace",
		Long:  "Inspect a namespace by either name or tenant-id",
		Example: `cli namespace inspect dev
cli namespace inspect --tenant-id 8f3c2e1a...
cli namespace inspect $(cli namespace ls -q | sed -n '2p')
cli namespace inspect --tenant-id $(cli namespace ls -q tenant-id | sed -n '2p')`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()

			tenantID, err := cmd.Flags().GetString("tenant-id")
			if err != nil {
				return err
			}

			if tenantID != "" {
				if _, err := uuid.Parse(tenantID); err != nil {
					return fmt.Errorf("invalid tenant ID: must be a valid UUID")
				}
			}

			if tenantID != "" && len(args) > 0 {
				return fmt.Errorf("cannot provide both a namespace name and --tenant-id")
			}

			if len(args) > 0 {
				if _, err := uuid.Parse(args[0]); err == nil {
					return fmt.Errorf("it looks like you provided a tenant ID; use the --tenant-id flag")
				}
			}

			if tenantID == "" && len(args) == 0 {
				return fmt.Errorf("please provide either a namespace name or --tenant-id")
			}

			resolver := services.NamespaceResolverName
			value := ""

			if tenantID != "" {
				resolver = services.NamespaceResolverTenantID
				value = tenantID
			} else {
				value = args[0]
			}

			ns, err := service.NamespaceResolve(cmd.Context(), resolver, value)
			if err != nil {
				return err
			}

			owner, err := service.UserResolve(cmd.Context(), ns.Owner)
			if err != nil {
				return err
			}

			totalDevices := ns.DevicesAcceptedCount +
				ns.DevicesPendingCount +
				ns.DevicesRejectedCount +
				ns.DevicesRemovedCount

			fmt.Fprintf(out, `Namespace:
  Name:        %s
  Type:        %s
  Owner:       %s
  Tenant ID:   %s
  Created At:  %s

Devices:
  Accepted:    %d
  Pending:     %d
  Rejected:    %d
  Removed:     %d
  Total:       %d

Members: %d
`,
				ns.Name,
				ns.Type,
				owner.Username,
				ns.TenantID,
				ns.CreatedAt.Format("2006-01-02"),
				ns.DevicesAcceptedCount,
				ns.DevicesPendingCount,
				ns.DevicesRejectedCount,
				ns.DevicesRemovedCount,
				totalDevices,
				len(ns.Members),
			)

			for _, m := range ns.Members {
				user, err := service.UserResolve(cmd.Context(), m.ID)
				if err != nil {
					return err
				}
				fmt.Fprintf(out, "  %-12s (%s)\n", user.Username, m.Role)
			}

			fmt.Fprintln(out, "\nLimits:")
			if ns.MaxDevices == -1 {
				fmt.Fprintln(out, "  Max Devices: unlimited")
			} else {
				fmt.Fprintf(out, "  Max Devices: %d\n", ns.MaxDevices)
			}

			return nil
		},
	}

	cmd.Flags().String("tenant-id", "", "Inspect namespace by tenant ID")

	return cmd
}

// memberCommands factory function that creates and returns a new command with
// add and remove subcommands dedicated to members management. It receives a service
// for handling business logic.
func memberCommands(service services.Services) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "member",
		Short: "Manage members",
		Long:  `Manage members`,
	}

	cmd.AddCommand(memberAdd(service))
	cmd.AddCommand(memberRemove(service))

	return cmd
}

func memberAdd(service services.Services) *cobra.Command {
	return &cobra.Command{
		Use:   "add <username> <namespace> <role>",
		Short: "Add a member",
		Long: `Adds a new member to the specified namespace with the given role. 
The username identifies the member to be added, the namespace specifies where the member should be added, 
and the role indicates the permissions that the member will have within that namespace.`,
		Example: `cli member add myuser mynamespace observer`,
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			role := authorizer.RoleFromString(args[2])
			if role == authorizer.RoleInvalid {
				return fmt.Errorf("invalid role %q, valid roles are: owner, administrator, operator, observer", args[2])
			}

			input := inputs.MemberAdd{
				Username:  args[0],
				Namespace: args[1],
				Role:      role,
			}

			namespace, err := service.NamespaceAddMember(cmd.Context(), &input)
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
	}
}

func memberRemove(service services.Services) *cobra.Command {
	return &cobra.Command{
		Use:   "remove <username> <namespace>",
		Short: "Remove a member",
		Long: `Removes an existing member from the specified namespace. 
The username identifies the member to be removed, and the namespace specifies where the member is currently located.`,
		Example: `cli member remove john_doe dev`,
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			input := inputs.MemberRemove{
				Username:  args[0],
				Namespace: args[1],
			}

			namespace, err := service.NamespaceRemoveMember(cmd.Context(), &input)
			if err != nil {
				return err
			}

			cmd.Println("Member removed successfully")
			cmd.Println("Namespace:", namespace.Name)
			cmd.Println("Tenant:", namespace.TenantID)
			cmd.Println("Member:", input.Username)

			return nil
		},
	}
}
