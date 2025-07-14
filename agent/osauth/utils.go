package osauth

type User struct {
	UID      uint32 // The user ID of the account.
	GID      uint32 // The group ID of the account.
	Username string // The login name of the account.
	Password string // The hashed password of the account.
	Name     string // The full name of the account owner.
	HomeDir  string // The home directory path of the account.
	Shell    string // The default login shell for the account.
}
