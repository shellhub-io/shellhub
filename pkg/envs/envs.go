package envs

// Backend is an interface for any sort of underlying key/value store.
type Backend interface {
	Get(key string) string
}

// DefaultBackend define the backend to be used to get environment variables.
var DefaultBackend Backend

func init() {
	DefaultBackend = &envBackend{}
}
