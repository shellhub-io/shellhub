package osauth

import (
	"os"
	"strings"
)

type User struct {
	UID      uint32 // The user ID of the account.
	GID      uint32 // The group ID of the account.
	Username string // The login name of the account.
	Password string // The hashed password of the account.
	Name     string // The full name of the account owner.
	HomeDir  string // The home directory path of the account.
	Shell    string // The default login shell for the account.
}

// PermitEmptyPasswords checks if the environment variable
// SHELLHUB_PERMIT_EMPTY_PASSWORDS is set to true. If so, it returns true,
// allowing empty passwords. Otherwise, it returns false.
func PermitEmptyPasswords() bool {
	// TODO: Consider reading this configuration from the main application's function
	// and passing it down to the osauth package.
	// TODO: Consider caching the result to avoid repeated environment variable lookups.
	return strings.EqualFold(os.Getenv("SHELLHUB_PERMIT_EMPTY_PASSWORDS"), "true")
}
