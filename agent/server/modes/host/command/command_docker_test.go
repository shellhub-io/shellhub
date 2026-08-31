//go:build docker

package command

import (
	"os"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNsenterCommandWrapper is an integration-level test for nsenterCommandWrapper.
// It injects statFn (to control which /proc/1/ns/* files appear present) and verifies:
//   - -T is never present in the assembled command (time namespace is never joined)
//   - namespace flags from statFn are included when present
//   - absent namespace flags are excluded
//   - the setpriv/nsenter prefix ordering is unchanged
//   - the user command appears at the end
//
//nolint:paralleltest
func TestNsenterCommandWrapper(t *testing.T) {
	origStatFn := statFn

	t.Cleanup(func() {
		statFn = origStatFn
	})

	presentNSFiles := map[string]bool{
		"/proc/1/ns/mnt": true,
		"/proc/1/ns/net": true,
		"/proc/1/ns/pid": true,
	}

	statFn = func(path string) (os.FileInfo, error) {
		if path == "/usr/bin/nsenter" {
			return nil, nil
		}

		if presentNSFiles[path] {
			return nil, nil
		}

		return nil, os.ErrNotExist
	}

	expectedNSFlags := []string{"-m", "-n", "-p"}

	t.Run("present ns flags included and -T never present", func(t *testing.T) {
		cmd, err := nsenterCommandWrapper(1000, 1000, []uint32{1000}, "/home/user", "/bin/sh")
		require.NoError(t, err)

		assert.NotContains(t, cmd, "-T")

		for _, flag := range expectedNSFlags {
			assert.Contains(t, cmd, flag)
		}

		for _, absent := range []string{"-u", "-i", "-C"} {
			assert.NotContains(t, cmd, absent)
		}

		setprivIdx := indexOf(cmd, "/bin/setpriv")
		nsenterIdx := indexOf(cmd, "/usr/bin/nsenter")
		assert.NotEqual(t, -1, setprivIdx, "/bin/setpriv must be present")
		assert.NotEqual(t, -1, nsenterIdx, "/usr/bin/nsenter must be present")
		assert.Less(t, setprivIdx, nsenterIdx, "/bin/setpriv must precede /usr/bin/nsenter")

		assert.Equal(t, "/bin/sh", cmd[len(cmd)-1])
	})

	t.Run("statFn controls which ns flags appear", func(t *testing.T) {
		statFn = func(path string) (os.FileInfo, error) {
			if path == "/usr/bin/nsenter" || path == "/proc/1/ns/net" {
				return nil, nil
			}

			return nil, os.ErrNotExist
		}

		cmd, err := nsenterCommandWrapper(1000, 1000, []uint32{1000}, "/home/user", "/bin/bash")
		require.NoError(t, err)

		assert.Contains(t, cmd, "-n")
		assert.NotContains(t, cmd, "-T")

		for _, absent := range []string{"-m", "-u", "-i", "-p", "-C"} {
			assert.NotContains(t, cmd, absent)
		}
	})

	t.Run("multiple ns flags from statFn — no -T", func(t *testing.T) {
		statFn = func(path string) (os.FileInfo, error) {
			if path == "/usr/bin/nsenter" || path == "/proc/1/ns/mnt" || path == "/proc/1/ns/uts" {
				return nil, nil
			}

			return nil, os.ErrNotExist
		}

		cmd, err := nsenterCommandWrapper(1000, 1000, []uint32{1000}, "/home/user", "/bin/bash")
		require.NoError(t, err)

		assert.Contains(t, cmd, "-m")
		assert.Contains(t, cmd, "-u")
		assert.NotContains(t, cmd, "-T")

		for _, absent := range []string{"-i", "-n", "-p", "-C"} {
			assert.NotContains(t, cmd, absent)
		}
	})
}

func indexOf(slice []string, target string) int {
	for i, s := range slice {
		if s == target {
			return i
		}
	}

	return -1
}

func TestNsenterArgs(t *testing.T) {
	t.Run("all flags in present are forwarded unchanged", func(t *testing.T) {
		present := map[string]string{
			"mnt":    "-m",
			"uts":    "-u",
			"ipc":    "-i",
			"net":    "-n",
			"pid":    "-p",
			"cgroup": "-C",
		}
		args := nsenterArgs(present)

		got := make([]string, len(args))
		copy(got, args)
		sort.Strings(got)

		expected := []string{"-C", "-i", "-m", "-n", "-p", "-u"}
		assert.Equal(t, expected, got)
	})

	t.Run("-T never appears regardless of input", func(t *testing.T) {
		present := map[string]string{
			"mnt": "-m",
			"net": "-n",
		}
		args := nsenterArgs(present)

		assert.NotContains(t, args, "-T")
		assert.Contains(t, args, "-m")
		assert.Contains(t, args, "-n")
	})

	t.Run("empty present returns empty slice", func(t *testing.T) {
		present := map[string]string{}
		args := nsenterArgs(present)

		assert.Empty(t, args)
	})
}
