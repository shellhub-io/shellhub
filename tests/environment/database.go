package environment

import "fmt"

// onlyPostgresAllowed panics if db is not "postgres". It documents at the
// call-site that mongo support has been removed and that postgres is the sole
// supported backend for integration tests.
func onlyPostgresAllowed(db string) {
	if db != "postgres" {
		panic(fmt.Sprintf("unsupported database %q: only postgres is supported", db))
	}
}
