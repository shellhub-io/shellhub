package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNginxController_generateConfigs_dirMode asserts that generateConfigs
// creates directories with permission bits 0o750 (owner rwx, group rx, other ---)
// rather than the less-restrictive 0o755.
func TestNginxController_generateConfigs_dirMode(t *testing.T) {
	// Arrange: create a temporary templates dir with a sub-directory and a
	// single dummy template file so filepath.Walk has something to visit.
	templatesDir := t.TempDir()
	subDir := filepath.Join(templatesDir, "conf.d")
	require.NoError(t, os.Mkdir(subDir, 0o750))

	// Write a template that expands to an empty config — no template actions
	// needed, just static text is fine for this test.
	tmplFile := filepath.Join(subDir, "test.conf")
	require.NoError(t, os.WriteFile(tmplFile, []byte("# test\n"), 0o600))

	// rootDir is where generated configs land.
	rootDir := t.TempDir()

	nc := &NginxController{
		rootDir:      rootDir,
		templatesDir: templatesDir,
		gatewayConfig: &GatewayConfig{
			Domain: "localhost",
		},
	}

	// Act
	nc.generateConfigs()

	// Assert: every directory created by generateConfigs under rootDir must
	// have mode 0o750.  We skip rootDir itself (created by t.TempDir(), not by
	// generateConfigs) and only check subdirectories that the code created.
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		require.NoError(t, err)

		if path == rootDir {
			return nil
		}

		if info.IsDir() {
			got := info.Mode().Perm()
			assert.Equal(t, os.FileMode(0o750), got,
				"directory %q has mode %04o, want 0750", path, got)
		}

		return nil
	})
	require.NoError(t, err)

	// Also verify that the hard-coded /etc/nginx creation path is exercised
	// by checking the generateConfigs helper uses 0o750.  Because we cannot
	// write to /etc/nginx in a test environment, we instead verify the
	// sub-directory under rootDir (conf.d) has the correct mode.
	subDirInRoot := filepath.Join(rootDir, "conf.d")
	info, err := os.Stat(subDirInRoot)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o750), info.Mode().Perm(),
		"conf.d directory mode should be 0750, got %04o", info.Mode().Perm())
}
