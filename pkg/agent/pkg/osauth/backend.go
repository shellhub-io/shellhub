package osauth

//go:generate mockery --name=Backend --filename=backend.go
type Backend interface {
	AuthUser(username, password string) bool
	LookupUser(username string) (*User, error)
}
