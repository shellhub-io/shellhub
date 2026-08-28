package agentd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadEnvFile(t *testing.T) {
	t.Run("returns nil when file does not exist", func(t *testing.T) {
		got := loadEnvFile("/no/such/file.env")
		assert.Empty(t, got)
	})

	t.Run("parses KEY=VALUE lines and skips blanks and comments", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "test.env")
		content := "FOO=bar\n# comment\n\nBAZ=qux\n"
		require.NoError(t, os.WriteFile(path, []byte(content), 0o600))

		got := loadEnvFile(path)
		assert.Equal(t, map[string]string{"FOO": "bar", "BAZ": "qux"}, got)
	})

	t.Run("preserves values containing equals signs", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "test.env")
		content := "URL=http://host:80/path?a=1&b=2\n"
		require.NoError(t, os.WriteFile(path, []byte(content), 0o600))

		got := loadEnvFile(path)
		assert.Equal(t, map[string]string{"URL": "http://host:80/path?a=1&b=2"}, got)
	})
}

func TestApplyEnvFileFallback(t *testing.T) {
	t.Run("sets missing vars from file without overriding existing ones", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "test.env")
		content := "TEST_ENVFILE_EXISTING=from-file\nTEST_ENVFILE_MISSING=from-file\n"
		require.NoError(t, os.WriteFile(path, []byte(content), 0o600))

		t.Setenv("TEST_ENVFILE_EXISTING", "from-env")

		applyEnvFileFallback(path)

		assert.Equal(t, "from-env", os.Getenv("TEST_ENVFILE_EXISTING"))
		assert.Equal(t, "from-file", os.Getenv("TEST_ENVFILE_MISSING"))
		t.Cleanup(func() { _ = os.Unsetenv("TEST_ENVFILE_MISSING") })
	})
}
