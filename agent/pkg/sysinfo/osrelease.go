package sysinfo

import (
	"bufio"
	"cmp"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	shellwords "github.com/mattn/go-shellwords"
)

// OSReleaseRoot is the directory [GetOSRelease] resolves etc/os-release and usr/lib/os-release
// under. The Docker agent points it at the host's root mount.
var OSReleaseRoot = "/"

var osReleaseFiles = []string{"etc/os-release", "usr/lib/os-release"}

// OSRelease is the distribution identity reported to the server.
type OSRelease struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// GetOSRelease reads the host's os-release file under [OSReleaseRoot]: etc/os-release, or
// usr/lib/os-release only when the first does not exist. ID defaults to "linux", and the name
// is PRETTY_NAME, then NAME, then "Linux". A line whose value does not parse is skipped. When
// neither file exists it returns those defaults and a nil error. It never returns a nil
// OSRelease: on an open or read error it returns the defaults alongside the error.
func GetOSRelease() (*OSRelease, error) {
	values, err := readOSRelease()

	return &OSRelease{
		ID:   cmp.Or(values["ID"], "linux"),
		Name: cmp.Or(values["PRETTY_NAME"], values["NAME"], "Linux"),
	}, err
}

func readOSRelease() (map[string]string, error) {
	for _, name := range osReleaseFiles {
		values, err := parseOSReleaseFile(filepath.Join(OSReleaseRoot, name))
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}

		return values, err
	}

	return map[string]string{}, nil
}

func parseOSReleaseFile(path string) (map[string]string, error) {
	file, err := os.Open(path) //nolint:gosec // path is one of the fixed os-release locations under OSReleaseRoot.
	if err != nil {
		return nil, err
	}
	defer file.Close() //nolint:errcheck // the file is only read, so a close error loses nothing.

	values := make(map[string]string)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		words, err := shellwords.Parse(value)
		if err != nil {
			continue
		}

		values[key] = strings.Join(words, " ")
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return values, nil
}
