package agentd

import (
	"bufio"
	"os"
	"strings"
)

const defaultEnvFilePath = "/etc/shellhub-agent.env"

// loadEnvFile parses a KEY=VALUE env file into a map, skipping blank lines and
// comments. Returns nil when the file does not exist or cannot be read.
func loadEnvFile(path string) map[string]string {
	f, err := os.Open(path) //nolint:gosec // path is defaultEnvFilePath in production, parameterised for tests.
	if err != nil {
		return nil
	}
	defer f.Close() //nolint:errcheck

	vars := make(map[string]string)

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		vars[key] = value
	}

	return vars
}

// applyEnvFileFallback loads the env file at path and sets any variable that is
// not already present in the process environment. Existing env vars are never
// overridden.
func applyEnvFileFallback(path string) {
	for key, value := range loadEnvFile(path) {
		if _, set := os.LookupEnv(key); !set {
			os.Setenv(key, value)
		}
	}
}
