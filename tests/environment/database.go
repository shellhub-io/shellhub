package environment

import "fmt"

func onlyPostgresAllowed(db string) {
	if db != "postgres" {
		panic(fmt.Sprintf("unsupported database %q: only postgres is supported", db))
	}
}
