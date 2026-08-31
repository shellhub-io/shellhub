package host

import (
	"bufio"
	"os"
	"slices"
	"strings"
)

var localeFiles = []string{
	"/etc/locale.conf",
	"/etc/default/locale",
	"/etc/environment",
}

const defaultLocale = "LANG=C.UTF-8"

func sessionEnv(clientEnv []string) []string {
	envs := systemLocaleEnv()
	if !hasCharacterLocale(envs) {
		envs = append([]string{defaultLocale}, envs...)
	}

	return append(envs, acceptClientEnv(clientEnv)...)
}

func systemLocaleEnv() []string {
	locale := make(map[string]string)

	for _, path := range localeFiles {
		readLocaleFile(path, locale)
	}

	envs := make([]string, 0, len(locale))
	for name, value := range locale {
		envs = append(envs, name+"="+value)
	}

	slices.Sort(envs)

	return envs
}

func readLocaleFile(path string, locale map[string]string) {
	file, err := os.Open(path) //nolint:gosec // the path is one of localeFiles, never input.
	if err != nil {
		return
	}

	defer file.Close() //nolint:errcheck

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		name, value, ok := parseLocaleLine(scanner.Text())
		if !ok {
			continue
		}

		if _, defined := locale[name]; !defined {
			locale[name] = value
		}
	}
}

func parseLocaleLine(line string) (string, string, bool) {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "#") {
		return "", "", false
	}

	name, value, ok := strings.Cut(line, "=")
	if !ok {
		return "", "", false
	}

	name = strings.TrimSpace(name)
	if name != "LANG" && !strings.HasPrefix(name, "LC_") {
		return "", "", false
	}

	value = strings.Trim(strings.TrimSpace(value), `"'`)
	if value == "" || strings.ContainsRune(value, 0) {
		return "", "", false
	}

	return name, value, true
}

func hasCharacterLocale(envs []string) bool {
	return slices.ContainsFunc(envs, func(env string) bool {
		name, _, _ := strings.Cut(env, "=")

		return name == "LANG" || name == "LC_ALL" || name == "LC_CTYPE"
	})
}

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
