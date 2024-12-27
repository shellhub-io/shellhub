package osauth

type User struct {
	Username string
	Password string
	Name     string
	HomeDir  string
	Shell    string
	UID      uint32
	GID      uint32
}
