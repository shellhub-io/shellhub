// Package testimage gives a test that builds a ShellHub Dockerfile the build arguments the
// Dockerfile declares without defaults: the versions pinned in versions.env at the repository
// root, the same file the compose wrapper and CI pass to their builds.
package testimage

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FileName is the file at the repository root that pins every build-time version.
const FileName = "versions.env"

// Root returns the repository root: the nearest directory at or above the working directory that
// holds versions.env. It returns an error when the working directory cannot be resolved or no
// directory above it holds the file, so a test never reads a version from the wrong checkout.
func Root() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for current := dir; ; current = filepath.Dir(current) {
		if _, err := os.Stat(filepath.Join(current, FileName)); err == nil {
			return current, nil
		}

		if filepath.Dir(current) == current {
			return "", fmt.Errorf("%s: no %s in this directory or any above it", dir, FileName)
		}
	}
}

// BuildArgs reads versions.env from root and returns its KEY=VALUE lines in the form
// testcontainers takes as build arguments. It returns an error when the file cannot be read or a
// line is not KEY=VALUE with both sides filled in, so a build never runs with an empty version.
func BuildArgs(root string) (map[string]*string, error) {
	path := filepath.Join(root, FileName)

	file, err := os.Open(path) //nolint:gosec // the caller names the repository root, and reading its versions file is the point
	if err != nil {
		return nil, err
	}

	defer file.Close() //nolint:errcheck // the file is read-only and fully consumed

	args := make(map[string]*string)
	scanner := bufio.NewScanner(file)

	for number := 1; scanner.Scan(); number++ {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok || key == "" || value == "" {
			return nil, fmt.Errorf("%s:%d: expected KEY=VALUE, got %q", path, number, line)
		}

		args[key] = &value
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return args, nil
}

// Version returns the value versions.env under root pins for key. It returns an error when the
// file cannot be read, has a line that is not KEY=VALUE or has no such key, so a caller never runs
// a tool at an unpinned version.
func Version(root, key string) (string, error) {
	args, err := BuildArgs(root)
	if err != nil {
		return "", err
	}

	value, ok := args[key]
	if !ok {
		return "", fmt.Errorf("%s: %s is not pinned", filepath.Join(root, FileName), key)
	}

	return *value, nil
}
