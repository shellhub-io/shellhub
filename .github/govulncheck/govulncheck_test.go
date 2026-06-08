package govulncheck

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// repoRoot returns the absolute path to the repository root, which is two
// directories above this package (.github/govulncheck/ → .github/ → root).
func repoRoot() string {
	return filepath.Join("..", "..")
}

// testdataPath returns the path to a fixture file under testdata/.
func testdataPath(name string) string {
	return filepath.Join("testdata", name)
}

func TestParseFindings_Called(t *testing.T) {
	f, err := os.Open(testdataPath("called.json"))
	require.NoError(t, err)

	defer f.Close()

	findings, err := ParseFindings(f)
	require.NoError(t, err)

	require.Len(t, findings, 1)
	assert.Equal(t, "GO-2024-0001", findings[0].OSVID)
	assert.True(t, findings[0].Called, "expected the finding to be marked as called")
}

func TestParseFindings_ImportedOnly(t *testing.T) {
	f, err := os.Open(testdataPath("imported_only.json"))
	require.NoError(t, err)

	defer f.Close()

	findings, err := ParseFindings(f)
	require.NoError(t, err)

	require.Len(t, findings, 1)
	assert.Equal(t, "GO-2024-0002", findings[0].OSVID)
	assert.False(t, findings[0].Called, "expected the finding to be marked as imported-only")
}

func TestParseFindings_DuplicateID(t *testing.T) {
	f, err := os.Open(testdataPath("duplicate_id.json"))
	require.NoError(t, err)

	defer f.Close()

	findings, err := ParseFindings(f)
	require.NoError(t, err)

	// Two findings with same OSV ID but different call sites — both returned.
	require.Len(t, findings, 2)

	for _, finding := range findings {
		assert.Equal(t, "GO-2024-0003", finding.OSVID)
		assert.True(t, finding.Called)
	}
}

func TestParseFindings_Empty(t *testing.T) {
	f, err := os.Open(testdataPath("empty.json"))
	require.NoError(t, err)

	defer f.Close()

	findings, err := ParseFindings(f)
	require.NoError(t, err)
	assert.Empty(t, findings)
}

func TestCheckFindings_SuppressCalledEntry(t *testing.T) {
	findings := []Finding{
		{OSVID: "GO-2024-0001", Called: true},
	}

	allow := []AllowEntry{
		{OSVID: "GO-2024-0001", Reachability: ReachabilityCalled},
	}

	var summary bytes.Buffer

	err := CheckFindings(findings, allow, &summary)
	assert.NoError(t, err, "called finding suppressed by matching allow entry should not fail")

	output := summary.String()
	assert.Contains(t, output, "GO-2024-0001", "suppressed vuln should appear in the step summary")
}

func TestCheckFindings_SuppressImportedOnlyEntry(t *testing.T) {
	findings := []Finding{
		{OSVID: "GO-2024-0002", Called: false},
	}

	allow := []AllowEntry{
		{OSVID: "GO-2024-0002", Reachability: ReachabilityImported},
	}

	var summary bytes.Buffer

	err := CheckFindings(findings, allow, &summary)
	assert.NoError(t, err, "imported-only finding suppressed by matching allow entry should not fail")

	output := summary.String()
	assert.Contains(t, output, "GO-2024-0002")
}

func TestCheckFindings_ImportedOnlyNotSuppressedByCalledEntry(t *testing.T) {
	// Allow entry says "called", but the finding is imported-only — should NOT suppress.
	findings := []Finding{
		{OSVID: "GO-2024-0002", Called: false},
	}

	allow := []AllowEntry{
		{OSVID: "GO-2024-0002", Reachability: ReachabilityCalled},
	}

	var summary bytes.Buffer

	err := CheckFindings(findings, allow, &summary)
	assert.Error(t, err, "imported-only finding not matching called allow entry should fail")
}

func TestCheckFindings_CalledNotSuppressedByImportedEntry(t *testing.T) {
	// Allow entry says "imported", but the finding is called — should NOT suppress.
	findings := []Finding{
		{OSVID: "GO-2024-0001", Called: true},
	}

	allow := []AllowEntry{
		{OSVID: "GO-2024-0001", Reachability: ReachabilityImported},
	}

	var summary bytes.Buffer

	err := CheckFindings(findings, allow, &summary)
	assert.Error(t, err, "called finding not matching imported allow entry should fail")
}

func TestCheckFindings_StaleAllowEntry(t *testing.T) {
	// No findings but allow list has an entry — stale, should fail.
	findings := []Finding{}

	allow := []AllowEntry{
		{OSVID: "GO-2024-9999", Reachability: ReachabilityCalled},
	}

	var summary bytes.Buffer

	err := CheckFindings(findings, allow, &summary)
	assert.Error(t, err, "stale allow entry should cause a failure")
}

func TestCheckFindings_UnsuppressedFindingFails(t *testing.T) {
	findings := []Finding{
		{OSVID: "GO-2024-0001", Called: true},
	}

	var summary bytes.Buffer

	err := CheckFindings(findings, nil, &summary)
	assert.Error(t, err, "unsuppressed finding should fail")
}

func TestCheckFindings_NoFindingsNoAllowEntries(t *testing.T) {
	var summary bytes.Buffer

	err := CheckFindings(nil, nil, &summary)
	assert.NoError(t, err)
}

func TestLoadAllowList(t *testing.T) {
	content := `# This is a comment
GO-2024-0001 called
GO-2024-0002 imported

# Another comment
GO-2024-0003 called
`

	tmpFile, err := os.CreateTemp(t.TempDir(), "allow*.txt")
	require.NoError(t, err)

	_, err = tmpFile.WriteString(content)
	require.NoError(t, err)
	require.NoError(t, tmpFile.Close())

	entries, err := LoadAllowList(tmpFile.Name())
	require.NoError(t, err)

	require.Len(t, entries, 3)
	assert.Equal(t, "GO-2024-0001", entries[0].OSVID)
	assert.Equal(t, ReachabilityCalled, entries[0].Reachability)
	assert.Equal(t, "GO-2024-0002", entries[1].OSVID)
	assert.Equal(t, ReachabilityImported, entries[1].Reachability)
	assert.Equal(t, "GO-2024-0003", entries[2].OSVID)
	assert.Equal(t, ReachabilityCalled, entries[2].Reachability)
}

func TestLoadAllowList_FileNotFound(t *testing.T) {
	_, err := LoadAllowList("/nonexistent/path/allow.txt")
	assert.Error(t, err)
}

func TestLoadAllowList_InvalidLine(t *testing.T) {
	content := "INVALID_LINE_WITHOUT_REACHABILITY\n"

	tmpFile, err := os.CreateTemp(t.TempDir(), "allow*.txt")
	require.NoError(t, err)

	_, err = tmpFile.WriteString(content)
	require.NoError(t, err)
	require.NoError(t, tmpFile.Close())

	_, err = LoadAllowList(tmpFile.Name())
	assert.Error(t, err)
}

func TestCheckFindings_DuplicateIDInAllowList(t *testing.T) {
	// Two findings with the same OSV ID but different call sites — allow entry should suppress both.
	findings := []Finding{
		{OSVID: "GO-2024-0003", Called: true},
		{OSVID: "GO-2024-0003", Called: true},
	}

	allow := []AllowEntry{
		{OSVID: "GO-2024-0003", Reachability: ReachabilityCalled},
	}

	var summary bytes.Buffer

	err := CheckFindings(findings, allow, &summary)
	assert.NoError(t, err, "single allow entry should suppress all findings with same OSV ID and same reachability")
}

func TestCheckFindings_SummaryContainsSuppressedEntries(t *testing.T) {
	findings := []Finding{
		{OSVID: "GO-2024-0001", Called: true},
		{OSVID: "GO-2024-0002", Called: false},
	}

	allow := []AllowEntry{
		{OSVID: "GO-2024-0001", Reachability: ReachabilityCalled},
		{OSVID: "GO-2024-0002", Reachability: ReachabilityImported},
	}

	var summary bytes.Buffer

	err := CheckFindings(findings, allow, &summary)
	require.NoError(t, err)

	output := summary.String()
	assert.True(t, strings.Contains(output, "GO-2024-0001"), "summary should contain suppressed vuln GO-2024-0001")
	assert.True(t, strings.Contains(output, "GO-2024-0002"), "summary should contain suppressed vuln GO-2024-0002")
}

// TestAllowlistBaselineFiles verifies that the security allowlist and scanner
// ignore files are present at the repository root with the expected format, and
// that CODEOWNERS references each of them.
func TestAllowlistBaselineFiles(t *testing.T) {
	root := repoRoot()

	t.Run("trivyignore exists and has header", func(t *testing.T) {
		path := filepath.Join(root, ".trivyignore")

		data, err := os.ReadFile(path) //nolint:gosec
		require.NoError(t, err, ".trivyignore must exist at repository root")

		require.True(t, strings.HasPrefix(string(data), "#"),
			".trivyignore must begin with a header comment (line starting with '#')")
	})

	t.Run("govulncheck allowlist exists and parses without error", func(t *testing.T) {
		path := filepath.Join(root, ".govulncheck-allow.txt")

		_, err := os.Stat(path)
		require.NoError(t, err, ".govulncheck-allow.txt must exist at repository root")

		entries, err := LoadAllowList(path)
		require.NoError(t, err, "LoadAllowList must parse .govulncheck-allow.txt without error")

		// The baseline file is intentionally empty (no active suppressions).
		assert.Empty(t, entries, "baseline .govulncheck-allow.txt should contain no active entries")
	})

	t.Run("semgrepignore exists and has header", func(t *testing.T) {
		path := filepath.Join(root, ".semgrepignore")

		data, err := os.ReadFile(path) //nolint:gosec
		require.NoError(t, err, ".semgrepignore must exist at repository root")

		require.True(t, strings.HasPrefix(string(data), "#"),
			".semgrepignore must begin with a header comment (line starting with '#')")
	})

	t.Run("CODEOWNERS references baseline security files", func(t *testing.T) {
		path := filepath.Join(root, ".github", "CODEOWNERS")

		f, err := os.Open(path) //nolint:gosec
		require.NoError(t, err, "CODEOWNERS must exist")

		defer f.Close() //nolint:errcheck

		var lines []string

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}

		require.NoError(t, scanner.Err())

		content := strings.Join(lines, "\n")

		for _, file := range []string{".trivyignore", ".govulncheck-allow.txt", ".semgrepignore"} {
			assert.True(t, strings.Contains(content, file),
				"CODEOWNERS must contain an entry for %s", file)
		}
	})
}
