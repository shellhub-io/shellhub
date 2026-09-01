package osauth

// Backend answers questions about the host's user accounts. It exists so that tests, and the
// container build, can substitute files other than the ones under /etc.
type Backend interface {
	AuthUser(username, password string) bool
	LookupUser(username string) (*User, error)
	ListGroups(username string) ([]uint32, error)
}
