package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadAllowlist(t *testing.T) {
	cases := []struct {
		description string
		content     string
		expected    map[string]string
		wantErr     bool
	}{
		{
			description: "reads path, symbol and reason",
			content:     "pkg/a.go KindInvalid zero value\n\npkg/b.go Backend mocked\n",
			expected: map[string]string{
				"shellhub/pkg/a.go KindInvalid": "shellhub/.deadcode-allow:1",
				"shellhub/pkg/b.go Backend":     "shellhub/.deadcode-allow:3",
			},
		},
		{
			description: "rejects an entry without a reason",
			content:     "pkg/a.go KindInvalid\n",
			wantErr:     true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			root := t.TempDir()
			require.NoError(t, os.WriteFile(filepath.Join(root, allowlistName), []byte(tc.content), 0o600))

			allowed := map[string]string{}
			err := readAllowlist(repository{name: "shellhub", root: root}, allowed)

			if tc.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.expected, allowed)
		})
	}
}

func TestReport(t *testing.T) {
	dead := finding{repo: "shellhub", path: "pkg/a.go", line: 3, symbol: "Unused", kind: "var"}

	cases := []struct {
		description string
		dead        map[string]finding
		allowed     map[string]string
		wantErr     bool
		wantLines   []string
	}{
		{
			description: "passes with nothing dead",
			dead:        map[string]finding{},
			allowed:     map[string]string{},
			wantLines:   []string{"no dead code"},
		},
		{
			description: "passes when every dead symbol is allowed",
			dead:        map[string]finding{dead.key(): dead},
			allowed:     map[string]string{dead.key(): "shellhub/.deadcode-allow:1"},
			wantLines:   []string{"no dead code"},
		},
		{
			description: "fails on a dead symbol nobody allowed, naming where it is",
			dead:        map[string]finding{dead.key(): dead},
			allowed:     map[string]string{},
			wantErr:     true,
			wantLines:   []string{"shellhub/pkg/a.go:3: unused var Unused"},
		},
		{
			description: "fails on an allowlist entry that is no longer dead, naming the entry",
			dead:        map[string]finding{},
			allowed:     map[string]string{dead.key(): "shellhub/.deadcode-allow:1"},
			wantErr:     true,
			wantLines:   []string{"shellhub/.deadcode-allow:1: shellhub/pkg/a.go Unused is used now, or gone; drop the entry"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			var out strings.Builder
			err := report(&out, findProblems(tc.dead, tc.allowed), 0)

			if tc.wantErr {
				require.ErrorIs(t, err, errDeadCode)
			} else {
				require.NoError(t, err)
			}

			for _, line := range tc.wantLines {
				assert.Contains(t, strings.Split(out.String(), "\n"), line)
			}
		})
	}
}

func TestSince(t *testing.T) {
	pending := finding{repo: "shellhub", path: "pkg/a.go", line: 3, symbol: "Pending", kind: "func"}
	moved := finding{repo: "shellhub", path: "pkg/a.go", line: 9, symbol: "Pending", kind: "func"}
	orphaned := finding{repo: "cloud", path: "pkg/b.go", line: 7, symbol: "Orphaned", kind: "func"}

	base := problems{
		dead:  map[string]finding{pending.key(): pending},
		stale: map[string]string{"shellhub/pkg/c.go Adopted": "shellhub/.deadcode-allow:1"},
	}
	head := problems{
		dead: map[string]finding{moved.key(): moved, orphaned.key(): orphaned},
		stale: map[string]string{
			"shellhub/pkg/c.go Adopted": "shellhub/.deadcode-allow:1",
			"cloud/pkg/d.go Revived":    "cloud/.deadcode-allow:2",
		},
	}

	introduced := head.since(base)

	assert.Equal(t, map[string]finding{orphaned.key(): orphaned}, introduced.dead, "a symbol dead at the base stays unreported after moving lines")
	assert.Equal(t, map[string]string{"cloud/pkg/d.go Revived": "cloud/.deadcode-allow:2"}, introduced.stale)

	var out strings.Builder
	require.ErrorIs(t, report(&out, introduced, head.count()-introduced.count()), errDeadCode)
	assert.Contains(t, strings.Split(out.String(), "\n"), "2 already at the base, not reported here")
}

func TestExportMergeBase(t *testing.T) {
	root := t.TempDir()
	git := func(args ...string) {
		t.Helper()

		cmd := exec.CommandContext(t.Context(), "git", append([]string{"-C", root, "-c", "user.name=test", "-c", "user.email=test@example.com"}, args...)...) //nolint:gosec // git on the test's own temporary repository
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, string(out))
	}
	write := func(name, content string) {
		t.Helper()

		require.NoError(t, os.MkdirAll(filepath.Join(root, filepath.Dir(name)), 0o750))
		require.NoError(t, os.WriteFile(filepath.Join(root, name), []byte(content), 0o600))
	}

	git("init", "-q", "-b", "trunk")
	write("a.go", "at base")
	write("pkg/b.go", "at base")
	require.NoError(t, os.Symlink("pkg/b.go", filepath.Join(root, "link.go")))
	git("add", ".")
	git("commit", "-q", "-m", "base")
	git("checkout", "-q", "-b", "change")
	write("a.go", "changed")
	write("c.go", "added by the change")
	git("add", ".")
	git("commit", "-q", "-m", "change")
	git("checkout", "-q", "trunk")
	write("d.go", "landed on trunk after the change branched")
	git("add", ".")
	git("commit", "-q", "-m", "later")
	git("checkout", "-q", "change")

	dest := filepath.Join(t.TempDir(), "shellhub")
	require.NoError(t, exportMergeBase(context.Background(), root, "trunk", dest))

	content, err := os.ReadFile(filepath.Join(dest, "a.go")) //nolint:gosec // reads back what the test just exported
	require.NoError(t, err)
	assert.Equal(t, "at base", string(content))
	assert.FileExists(t, filepath.Join(dest, "pkg", "b.go"))
	assert.NoFileExists(t, filepath.Join(dest, "c.go"), "the change's own files are not in its base")
	assert.NoFileExists(t, filepath.Join(dest, "d.go"), "the base is the merge base, not the ref's tip")

	link, err := os.Readlink(filepath.Join(dest, "link.go"))
	require.NoError(t, err)
	assert.Equal(t, "pkg/b.go", link)
}

func TestDeadUnderEvery(t *testing.T) {
	both := finding{repo: "shellhub", path: "a.go", symbol: "Both"}
	dockerOnly := finding{repo: "shellhub", path: "b.go", symbol: "DockerOnly"}

	collected := map[string]map[string]finding{
		"docker": {both.key(): both, dockerOnly.key(): dockerOnly},
		"native": {both.key(): both},
	}

	dead, err := deadUnderEvery([]string{"docker", "native"}, func(tags string) (map[string]finding, error) {
		return collected[tags], nil
	})

	require.NoError(t, err)
	assert.Equal(t, map[string]finding{both.key(): both}, dead)
}

func TestMainCheckout(t *testing.T) {
	cases := []struct {
		description string
		dotGit      func(t *testing.T, root string)
		expected    func(root string) string
	}{
		{
			description: "returns a main checkout itself",
			dotGit: func(t *testing.T, root string) {
				t.Helper()
				require.NoError(t, os.Mkdir(filepath.Join(root, ".git"), 0o700))
			},
			expected: func(root string) string { return root },
		},
		{
			description: "follows a linked worktree to its main checkout",
			dotGit: func(t *testing.T, root string) {
				t.Helper()
				require.NoError(t, os.WriteFile(filepath.Join(root, ".git"), []byte("gitdir: /src/shellhub/.git/worktrees/x\n"), 0o600))
			},
			expected: func(string) string { return "/src/shellhub" },
		},
		{
			description: "resolves a relative gitdir against the worktree",
			dotGit: func(t *testing.T, root string) {
				t.Helper()
				require.NoError(t, os.WriteFile(filepath.Join(root, ".git"), []byte("gitdir: ../main/.git/worktrees/x\n"), 0o600))
			},
			expected: func(root string) string { return filepath.Join(filepath.Dir(root), "main") },
		},
		{
			description: "returns a submodule checkout itself",
			dotGit: func(t *testing.T, root string) {
				t.Helper()
				require.NoError(t, os.WriteFile(filepath.Join(root, ".git"), []byte("gitdir: ../.git/modules/shellhub\n"), 0o600))
			},
			expected: func(root string) string { return root },
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			root := filepath.Join(t.TempDir(), "checkout")
			require.NoError(t, os.Mkdir(root, 0o700))
			tc.dotGit(t, root)

			assert.Equal(t, tc.expected(root), mainCheckout(root))
		})
	}
}

func TestUnusedDeclarations(t *testing.T) {
	root := t.TempDir()
	files := map[string]string{
		"go.mod": "module github.com/shellhub-io/fixture\n\ngo 1.22\n",
		"status/status.go": `package status

type Status string

const (
	Active   Status = "active"
	Canceled Status = "canceled"
)

func Parse(s string) Status { return Status(s) }
`,
		"phase/phase.go": `package phase

type Phase string

const (
	PhaseOpen Phase = "open"
	PhaseVoid Phase = "void"
)

func IsOpen(s string) bool { return s == string(PhaseOpen) }
`,
		"kind/kind.go": `package kind

type Kind int

const (
	KindA Kind = iota
	KindB
)
`,
		"backend/backend.go": `package backend

type Backend interface{ Call() }

type Orphan interface{ Run() }
`,
		"backend/mocks/mock_backend.go": `package mocks

type MockBackend struct{}

func (MockBackend) Call() {}
`,
	}
	for name, content := range files {
		require.NoError(t, os.MkdirAll(filepath.Join(root, filepath.Dir(name)), 0o750))
		require.NoError(t, os.WriteFile(filepath.Join(root, name), []byte(content), 0o600))
	}

	found, err := unusedDeclarations(t.Context(), root, []repository{{name: "shellhub", root: root}}, "off", "")
	require.NoError(t, err)

	var symbols []string
	for _, f := range found {
		symbols = append(symbols, f.path+" "+f.symbol)
	}

	assert.ElementsMatch(t, []string{
		"kind/kind.go KindA",
		"kind/kind.go KindB",
		"backend/backend.go Orphan",
	}, symbols, "an enum member is live while its type or a sibling is used, and a mocked interface is live")
}
