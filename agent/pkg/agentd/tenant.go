package agentd

import (
	"os"
	"path/filepath"
	"strings"
)

// TenantFilePath returns the path where the agent persists the tenant learned
// from a pairing. It is a sibling of the private key so it lands on the same
// persistent mount, and suffixing the key name avoids collisions when multiple
// agents share a directory with different keys.
func TenantFilePath(privateKeyPath string) string {
	return privateKeyPath + ".tenant"
}

// ReadPersistedTenant reads the tenant persisted by a previous pairing. A
// missing file is not an error; it returns an empty tenant.
func ReadPersistedTenant(path string) (string, error) {
	data, err := os.ReadFile(path) //nolint:gosec // path is derived from the PRIVATE_KEY env var (operator-configured, not end-user input).
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}

		return "", err
	}

	return strings.TrimSpace(string(data)), nil
}

// PersistTenant atomically writes the tenant learned from a pairing. The file
// is plain text (one line) so an operator can inspect or write it by hand as
// an escape hatch.
func PersistTenant(path, tenant string) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())

	if err := tmp.Chmod(0o600); err != nil {
		tmp.Close()

		return err
	}

	if _, err := tmp.WriteString(tenant + "\n"); err != nil {
		tmp.Close()

		return err
	}

	if err := tmp.Sync(); err != nil {
		tmp.Close()

		return err
	}

	if err := tmp.Close(); err != nil {
		return err
	}

	return os.Rename(tmp.Name(), path)
}
