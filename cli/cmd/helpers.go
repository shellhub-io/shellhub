package cmd

import (
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/shellhub-io/shellhub/pkg/models"
)

// helper to print the listing table
func printUsersTable(out io.Writer, users []models.User) {
	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "USERNAME\tEMAIL\tROLE")
	for _, u := range users {
		role := "user"
		if u.Admin {
			role = "admin"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\n", u.Username, u.Email, role)
	}
	w.Flush()
}
