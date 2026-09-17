package environment

import "fmt"

func onlyPostgresAllowed(db string) error {
	if db != "postgres" {
		return fmt.Errorf("unsupported database %q: only postgres is supported", db)
	}

	return nil
}
