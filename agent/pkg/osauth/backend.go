package osauth

// Backend answers questions about the host's user accounts. It exists so that tests, and the
// container build, can substitute files other than the ones under /etc.
type Backend interface {
	AuthUser(username, password string) bool
	LookupUser(username string) (*User, error)
	AccountExpired(username string) bool
	ListGroups(username string) ([]uint32, error)
}

// Set serves AuthUser, LookupUser, AccountExpired and ListGroups from b until t and its subtests
// finish, then puts back the backend it replaced. Calls nest: an inner Set gives back the outer one
// when its test ends. The backend is process-global, so a test that calls Set must not run in
// parallel with one that reads the host's accounts.
func Set(t interface{ Cleanup(func()) }, b Backend) {
	previous := DefaultBackend
	t.Cleanup(func() { DefaultBackend = previous })
	DefaultBackend = b
}
