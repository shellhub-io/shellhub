package host

import (
	"strings"
)

// acceptClientEnv filters the environment variables forwarded by an SSH client, keeping only those
// that are safe to pass into a session. It mirrors the OpenSSH default AcceptEnv behaviour of
// accepting LANG and any variable whose name starts with "LC_". All other variables — including
// dangerous loader/shell overrides (LD_*, BASH_ENV, GODEBUG, PATH, HOME, USER, TERM,
// SSH_AUTH_SOCK, ...) — are silently dropped.
//
// Entries that are malformed (no "=" separator), have an empty name, or contain a NUL byte are
// also dropped. See advisory GHSA-22pj-m7g6-rh43.
func acceptClientEnv(envs []string) []string {
	if envs == nil {
		return nil
	}

	result := make([]string, 0, len(envs))

	for _, e := range envs {
		if strings.ContainsRune(e, 0) {
			continue
		}

		name, _, ok := strings.Cut(e, "=")
		if !ok || name == "" {
			continue
		}

		if name == "LANG" || strings.HasPrefix(name, "LC_") {
			result = append(result, e)
		}
	}

	return result
}
