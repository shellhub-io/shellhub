package envs

// Backend is an interface for any sort of underlying key/value store
type Backend interface {
	Get(key string) string
}

// Get value from environment variables
var DefaultBackend Backend

func init() {
	DefaultBackend = &envBackend{}
}
