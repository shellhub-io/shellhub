package host

import (
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestOpenPty_TtyNameSafeAfterClose verifies the invariant documented in the
// comment above `defer tty.Close()` in openPty: the tty file descriptor is
// closed before openPty returns, but callers only ever use tty.Name() — a
// path-based lookup — not the underlying fd.  Therefore os.Chown (and any
// other path-based syscall) remains safe even though the fd is already closed.
//
// Concretely the test asserts:
//  1. openPty succeeds and returns a non-nil tty whose Name() is non-empty.
//  2. The path returned by tty.Name() is still stat-able after openPty returns
//     (i.e. the path is valid independently of the fd's lifetime).
//  3. The tty fd itself is closed when openPty returns (read returns an error).
func TestOpenPty_TtyNameSafeAfterClose(t *testing.T) {
	// Use a real command that exits quickly so openPty can complete.
	c := exec.Command("/bin/true")

	ptmx, tty, err := openPty(c)
	require.NoError(t, err, "openPty must succeed in this environment")

	t.Cleanup(func() {
		_ = ptmx.Close()
		_ = c.Wait()
	})

	// 1. tty must be non-nil and carry a valid path.
	require.NotNil(t, tty, "openPty must return a non-nil tty")

	name := tty.Name()
	assert.NotEmpty(t, name, "tty.Name() must return a non-empty path")

	// 2. The path itself must still be valid on the filesystem (stat succeeds).
	// Note: the tty device node may disappear on some kernels after the process
	// using the pty exits, but it always exists while the ptmx fd is open.
	_, statErr := os.Stat(name)
	assert.NoError(t, statErr, "os.Stat(tty.Name()) must succeed while ptmx is open")

	// 3. The fd returned in the tty *os.File is closed by openPty's defer, so
	// any fd-based operation must fail.  We use Read as a lightweight probe.
	buf := make([]byte, 1)
	_, readErr := tty.Read(buf)
	assert.Error(t, readErr, "tty fd must be closed after openPty returns")
}
