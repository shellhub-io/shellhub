package main_test

// mocks_structural_gates_test.go validates all structural gates for the
// mockery v3 migration: Dockerfile cleanliness, absence of go:generate
// directives, extended wrong-sibling checks, and devscripts/gen-mock purity.
//
// Verification criteria:
//  1. No Dockerfile (available in the api container) references mockery/v2.
//  2. No .go source file in api/ or pkg/ contains a //go:generate mockery
//     directive (excludes test files to avoid self-referential false positives).
//  3. Extended wrong sibling files are absent: pkg/envs/mocks/backend.go,
//     pkg/api/internalclient/mocks/client.go, pkg/uuid/mocks/u_u_i_d.go.
//  4. devscripts/gen-mock contains no v2-style per-interface flags; skipped
//     when the file is not present in the current container mount.
//  5. Every mock header references the v3 source URL (github.com/vektra/mockery
//     without the /v2 suffix).

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// dockerfilesInContainer returns Dockerfiles that are mounted in the api
// container (relative to repoRoot). It checks for existence before returning
// so callers do not need to guard against missing paths.
func dockerfilesInContainer(t *testing.T) []string {
	t.Helper()

	root := repoRoot(t)

	candidates := []string{
		"api/Dockerfile",
		"openapi/Dockerfile",
	}

	var found []string

	for _, rel := range candidates {
		abs := filepath.Join(root, rel)
		if _, err := os.Stat(abs); err == nil {
			found = append(found, rel)
		}
	}

	return found
}

// readFileLines opens path and returns all lines.
func readFileLines(t *testing.T, path string) []string {
	t.Helper()

	f, err := os.Open(path) //nolint:gosec
	require.NoError(t, err, "open %s", path)

	defer f.Close() //nolint:errcheck

	var lines []string
	sc := bufio.NewScanner(f)

	for sc.Scan() {
		lines = append(lines, sc.Text())
	}

	require.NoError(t, sc.Err())

	return lines
}

// min returns the smaller of a and b.
func min(a, b int) int {
	if a < b {
		return a
	}

	return b
}

// walkSourceGoFiles calls fn for every non-test *.go file found under root,
// skipping vendor and mocks directories.
func walkSourceGoFiles(t *testing.T, root string, fn func(path string)) {
	t.Helper()

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			switch info.Name() {
			case "vendor", "mocks":
				return filepath.SkipDir
			}
		}

		isSourceFile := !info.IsDir() &&
			strings.HasSuffix(path, ".go") &&
			!strings.HasSuffix(path, "_test.go")

		if isSourceFile {
			fn(path)
		}

		return nil
	})

	require.NoError(t, err)
}

// TestStructuralGates_MockHeaderV3Source verifies that every generated mock
// header references the v3 source URL (github.com/vektra/mockery without the
// /v2 suffix) and does NOT reference the v2 URL.
func TestStructuralGates_MockHeaderV3Source(t *testing.T) {
	root := repoRoot(t)

	// mockery v3 emits "// github.com/vektra/mockery" (no "/v2" suffix).
	// v2 emits "// github.com/vektra/mockery/v2".
	const wantFragment = "github.com/vektra/mockery"
	const badFragment = "github.com/vektra/mockery/v2"

	for _, rel := range mockFiles() {
		abs := filepath.Join(root, rel)

		t.Run(rel, func(t *testing.T) {
			lines := readFileLines(t, abs)
			require.NotEmpty(t, lines, "mock file must not be empty: %s", rel)

			found := false

			for _, line := range lines[:min(len(lines), 5)] {
				if strings.Contains(line, wantFragment) {
					found = true
				}

				assert.NotContains(t, line, badFragment,
					"mock file %s header must not reference the v2 URL", rel)
			}

			assert.True(t, found,
				"mock file %s header must contain %q in first 5 lines", rel, wantFragment)
		})
	}
}

// TestStructuralGates_NoV2InDockerfiles verifies that no Dockerfile available
// in the api container installs or references mockery/v2.
func TestStructuralGates_NoV2InDockerfiles(t *testing.T) {
	for _, rel := range dockerfilesInContainer(t) {
		t.Run(rel, func(t *testing.T) {
			abs := filepath.Join(repoRoot(t), rel)
			lines := readFileLines(t, abs)

			for i, line := range lines {
				assert.NotContains(t, line, "mockery/v2",
					"Dockerfile %s line %d must not reference mockery/v2", rel, i+1)
			}
		})
	}
}

// TestStructuralGates_NoGoGenerateMockery verifies that no non-test .go source
// file in the api/ or pkg/ trees retains a //go:generate mockery directive.
// All generation is now driven by devscripts/gen-mock + .mockery.yaml.
func TestStructuralGates_NoGoGenerateMockery(t *testing.T) {
	root := repoRoot(t)
	const directive = "//go:generate mockery"

	for _, subtree := range []string{"api", "pkg"} {
		subtreeAbs := filepath.Join(root, subtree)

		if _, err := os.Stat(subtreeAbs); os.IsNotExist(err) {
			continue
		}

		walkSourceGoFiles(t, subtreeAbs, func(path string) {
			lines := readFileLines(t, path)

			for i, line := range lines {
				assert.NotContains(t, line, directive,
					"file %s line %d must not contain '%s'", path, i+1, directive)
			}
		})
	}
}

// TestStructuralGates_ExtendedWrongSiblingFilesAbsent verifies that the
// additional wrong-sibling files that a misconfigured v2 run would produce
// are absent from the repository.
func TestStructuralGates_ExtendedWrongSiblingFilesAbsent(t *testing.T) {
	root := repoRoot(t)

	wrongFiles := []string{
		"pkg/envs/mocks/backend.go",
		"pkg/api/internalclient/mocks/client.go",
		"pkg/uuid/mocks/u_u_i_d.go",
	}

	for _, rel := range wrongFiles {
		abs := filepath.Join(root, rel)
		_, err := os.Stat(abs)
		assert.True(t, os.IsNotExist(err),
			"wrong sibling file must not exist: %s", rel)
	}
}

// TestStructuralGates_GenMockNoV2Flags verifies that devscripts/gen-mock
// contains no v2-style per-interface CLI flags or references to mockery/v2.
// The test is skipped when devscripts/ is not mounted in the current container.
func TestStructuralGates_GenMockNoV2Flags(t *testing.T) {
	root := repoRoot(t)
	abs := filepath.Join(root, "devscripts/gen-mock")

	if _, err := os.Stat(abs); os.IsNotExist(err) {
		t.Skip("devscripts/gen-mock not present in this container mount — skipping")
	}

	lines := readFileLines(t, abs)

	v2Indicators := []string{
		"mockery/v2",
		"--name ",
		"--output ",
		"--structname ",
		"--filename ",
		"--print",
		"--all ",
	}

	for i, line := range lines {
		for _, indicator := range v2Indicators {
			assert.NotContains(t, line, indicator,
				"devscripts/gen-mock line %d must not contain v2 flag/reference %q", i+1, indicator)
		}
	}
}
