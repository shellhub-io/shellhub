package host

import (
	"bufio"
	"os"
	"slices"
	"strings"
)

// localeFiles are the system wide files a locale may be configured in, in the order they are
// consulted. The first file to define a variable wins, so a systemd locale.conf takes precedence
// over the Debian defaults. Under the docker build tag these point into the mounted host root.
var localeFiles = []string{
	"/etc/locale.conf",
	"/etc/default/locale",
	"/etc/environment",
}

// defaultLocale is used when the host configures no character locale at all, which is the common
// state on embedded targets. Without it those sessions run under the C locale, where tools writing
// to a TTY escape every byte >= 0x80 instead of printing the character.
//
// C.UTF-8 needs no generated locale data, unlike a language specific locale that glibc would fall
// back to C for when absent.
const defaultLocale = "LANG=C.UTF-8"

// sessionEnv builds the locale environment a session starts with.
//
// The agent assembles the session environment from scratch, so nothing of the host's own
// configuration reaches the shell unless it is read here. A regular login gets it through
// pam_env; a ShellHub session would otherwise run under the C locale even on a host that
// configures a locale.
//
// The client's own variables come last: os/exec keeps the last value of a repeated key, so a
// client forwarding LANG still overrides the host.
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
