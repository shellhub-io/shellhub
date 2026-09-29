package main

import (
	"os"
	"path/filepath"
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
	}{
		{
			description: "passes with nothing dead",
			dead:        map[string]finding{},
			allowed:     map[string]string{},
		},
		{
			description: "passes when every dead symbol is allowed",
			dead:        map[string]finding{dead.key(): dead},
			allowed:     map[string]string{dead.key(): "shellhub/.deadcode-allow:1"},
		},
		{
			description: "fails on a dead symbol nobody allowed",
			dead:        map[string]finding{dead.key(): dead},
			allowed:     map[string]string{},
			wantErr:     true,
		},
		{
			description: "fails on an allowlist entry that is no longer dead",
			dead:        map[string]finding{},
			allowed:     map[string]string{dead.key(): "shellhub/.deadcode-allow:1"},
			wantErr:     true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			err := report(tc.dead, tc.allowed)

			if tc.wantErr {
				assert.ErrorIs(t, err, errDeadCode)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestIntersect(t *testing.T) {
	a := finding{repo: "shellhub", path: "a.go", symbol: "A"}
	b := finding{repo: "shellhub", path: "b.go", symbol: "B"}

	result := intersect(nil, map[string]finding{a.key(): a, b.key(): b})
	result = intersect(result, map[string]finding{a.key(): a})

	assert.Equal(t, map[string]finding{a.key(): a}, result)
}
