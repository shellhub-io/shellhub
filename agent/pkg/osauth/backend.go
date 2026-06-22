package osauth

type Backend interface {
	AuthUser(username, password string) bool
	LookupUser(username string) (*User, error)
	ListGroups(username string) ([]uint32, error)
}
