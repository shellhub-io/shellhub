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

// IsEnterprise returns true if the current ShellHub server instance is enterprise.
func IsEnterprise() bool {
	return DefaultBackend.Get("SHELLHUB_ENTERPRISE") == "true"
}

// IsCloud returns true if the current ShellHub server instance is cloud.
func IsCloud() bool {
	return DefaultBackend.Get("SHELLHUB_CLOUD") == "true"
}
