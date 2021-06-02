package envs

import "os"

type envBackend struct {
}

// envBackend is the default key/value store that reads from environment variables
func (b *envBackend) Get(name string) string {
	return os.Getenv(name)
}

// IsEnterprise returns true if the current ShellHub server instance is enterprise
func IsEnterprise() bool {
	return DefaultBackend.Get("SHELLHUB_ENTERPRISE") == "true"
}
