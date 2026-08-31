package agentd

import (
	"bufio"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

const defaultEnvFilePath = "/etc/shellhub-agent.env"

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

func applyEnvFileFallback(path string) {
	for key, value := range loadEnvFile(path) {
		if _, set := os.LookupEnv(key); !set {
			if err := os.Setenv(key, value); err != nil {
				log.WithError(err).WithField("key", key).Warning("failed to apply env file fallback")
			}
		}
	}
}
