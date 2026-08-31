package host

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-playground/assert/v2"
)

func TestAcceptClientEnv(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "accept LANG",
			input:    []string{"LANG=en_US.UTF-8"},
			expected: []string{"LANG=en_US.UTF-8"},
		},
		{
			name:     "accept LC_ALL",
			input:    []string{"LC_ALL=C"},
			expected: []string{"LC_ALL=C"},
		},
		{
			name:     "accept LC_CTYPE",
			input:    []string{"LC_CTYPE=UTF-8"},
			expected: []string{"LC_CTYPE=UTF-8"},
		},
		{
			name:     "accept LC_PAPER",
			input:    []string{"LC_PAPER=x"},
			expected: []string{"LC_PAPER=x"},
		},
		{
			name:     "accept LC_ with empty suffix (LC_=value)",
			input:    []string{"LC_="},
			expected: []string{"LC_="},
		},
		{
			name:     "drop LD_PRELOAD",
			input:    []string{"LD_PRELOAD=/"},
			expected: []string{},
		},
		{
			name:     "drop LD_AUDIT",
			input:    []string{"LD_AUDIT=/"},
			expected: []string{},
		},
		{
			name:     "drop GODEBUG",
			input:    []string{"GODEBUG=inittrace=1"},
			expected: []string{},
		},
		{
			name:     "drop PATH",
			input:    []string{"PATH=/tmp/evil"},
			expected: []string{},
		},
		{
			name:     "drop BASH_ENV",
			input:    []string{"BASH_ENV=/tmp/x"},
			expected: []string{},
		},
		{
			name:     "drop TERM",
			input:    []string{"TERM=evil"},
			expected: []string{},
		},
		{
			name:     "drop HOME",
			input:    []string{"HOME=/tmp"},
			expected: []string{},
		},
		{
			name:     "drop USER",
			input:    []string{"USER=root"},
			expected: []string{},
		},
		{
			name:     "drop SSH_AUTH_SOCK",
			input:    []string{"SSH_AUTH_SOCK=/tmp/x"},
			expected: []string{},
		},
		{
			name:     "drop arbitrary variable MYVAR",
			input:    []string{"MYVAR=value"},
			expected: []string{},
		},
		{
			name:     "drop LANGUAGE (not LC_ prefix match)",
			input:    []string{"LANGUAGE=en"},
			expected: []string{},
		},
		{
			name:     "drop LCALL (not LC_ prefix)",
			input:    []string{"LCALL=x"},
			expected: []string{},
		},
		{
			name:     "drop LD_ prefix variable",
			input:    []string{"LD_=x"},
			expected: []string{},
		},
		{
			name:     "drop malformed entry without equals sign",
			input:    []string{"FOO"},
			expected: []string{},
		},
		{
			name:     "drop empty-name entry (=value)",
			input:    []string{"=value"},
			expected: []string{},
		},
		{
			name:     "drop entry with NUL byte",
			input:    []string{"LANG=en\x00evil"},
			expected: []string{},
		},
		{
			name:     "nil input returns nil",
			input:    nil,
			expected: nil,
		},
		{
			name:     "empty input returns empty",
			input:    []string{},
			expected: []string{},
		},
		{
			name:     "order preserved for multiple accepted entries",
			input:    []string{"LC_ALL=C", "MYVAR=bad", "LANG=en_US.UTF-8", "LD_PRELOAD=/", "LC_CTYPE=UTF-8"},
			expected: []string{"LC_ALL=C", "LANG=en_US.UTF-8", "LC_CTYPE=UTF-8"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := acceptClientEnv(tt.input)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func withLocaleFiles(t *testing.T, contents ...string) {
	t.Helper()

	dir := t.TempDir()
	original := localeFiles

	t.Cleanup(func() { localeFiles = original })

	paths := make([]string, 0, len(contents))

	for i, content := range contents {
		path := filepath.Join(dir, fmt.Sprintf("locale%d", i))

		if content != "" {
			if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
				t.Fatalf("failed to write the locale file: %v", err)
			}
		}

		paths = append(paths, path)
	}

	localeFiles = paths
}

func TestSystemLocaleEnv(t *testing.T) {
	tests := []struct {
		name     string
		files    []string
		expected []string
	}{
		{
			name:     "no locale file on the host",
			files:    []string{"", "", ""},
			expected: []string{},
		},
		{
			name:     "reads a configured locale",
			files:    []string{"LANG=pt_BR.UTF-8\n", "", ""},
			expected: []string{"LANG=pt_BR.UTF-8"},
		},
		{
			name:     "skips comments and blank lines",
			files:    []string{"# the system locale\n\nLANG=en_US.UTF-8\n", "", ""},
			expected: []string{"LANG=en_US.UTF-8"},
		},
		{
			name:     "unquotes a quoted value",
			files:    []string{"LANG=\"pt_BR.UTF-8\"\n", "", ""},
			expected: []string{"LANG=pt_BR.UTF-8"},
		},
		{
			name:     "drops variables outside the allowlist",
			files:    []string{"PATH=/evil\nLD_PRELOAD=/evil.so\nLANG=C.UTF-8\n", "", ""},
			expected: []string{"LANG=C.UTF-8"},
		},
		{
			name:     "drops a value carrying a NUL byte",
			files:    []string{"LANG=en\x00evil\n", "", ""},
			expected: []string{},
		},
		{
			name:     "the first file to define a variable wins",
			files:    []string{"LANG=pt_BR.UTF-8\n", "LANG=en_US.UTF-8\n", ""},
			expected: []string{"LANG=pt_BR.UTF-8"},
		},
		{
			name:     "merges variables across files",
			files:    []string{"LANG=pt_BR.UTF-8\n", "LC_TIME=en_US.UTF-8\n", ""},
			expected: []string{"LANG=pt_BR.UTF-8", "LC_TIME=en_US.UTF-8"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withLocaleFiles(t, tt.files...)

			assert.Equal(t, tt.expected, systemLocaleEnv())
		})
	}
}

func TestSessionEnv(t *testing.T) {
	tests := []struct {
		name      string
		files     []string
		clientEnv []string
		expected  []string
	}{
		{
			name:      "falls back when the host configures no locale",
			files:     []string{"", "", ""},
			clientEnv: nil,
			expected:  []string{"LANG=C.UTF-8"},
		},
		{
			name:      "keeps the host locale over the fallback",
			files:     []string{"LANG=pt_BR.UTF-8\n", "", ""},
			clientEnv: nil,
			expected:  []string{"LANG=pt_BR.UTF-8"},
		},
		{
			name:      "falls back when the host sets no character locale",
			files:     []string{"LC_TIME=pt_BR.UTF-8\n", "", ""},
			clientEnv: nil,
			expected:  []string{"LANG=C.UTF-8", "LC_TIME=pt_BR.UTF-8"},
		},
		{
			name:      "the client comes last so it overrides the host",
			files:     []string{"LANG=pt_BR.UTF-8\n", "", ""},
			clientEnv: []string{"LANG=ja_JP.UTF-8", "PATH=/evil"},
			expected:  []string{"LANG=pt_BR.UTF-8", "LANG=ja_JP.UTF-8"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withLocaleFiles(t, tt.files...)

			assert.Equal(t, tt.expected, sessionEnv(tt.clientEnv))
		})
	}
}
