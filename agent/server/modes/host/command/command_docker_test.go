//go:build docker
// +build docker

package command

import (
	"os"
	"sort"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNsenterCommandWrapper is an integration-level test for nsenterCommandWrapper.
// It injects statFn (to control which /proc/1/ns/* files appear present) and
// timeNSJoinableFn (to control the time-namespace probe outcome) and verifies:
//   - when timeNSJoinableFn returns true, -T is present in the assembled command
//   - when timeNSJoinableFn returns false, -T is absent
//   - namespace flags from statFn are included regardless of the time-ns outcome
//   - the setpriv/nsenter prefix ordering is unchanged
//
//nolint:paralleltest
func TestNsenterCommandWrapper(t *testing.T) {
	origStatFn := statFn
	origTimeNSJoinableFn := timeNSJoinableFn

	t.Cleanup(func() {
		statFn = origStatFn
		timeNSJoinableFn = origTimeNSJoinableFn
	})

	// presentNSFiles is the set of /proc/1/ns/* files that statFn will report as present.
	// We choose a deterministic subset: mnt, net, pid.
	presentNSFiles := map[string]bool{
		"/proc/1/ns/mnt": true,
		"/proc/1/ns/net": true,
		"/proc/1/ns/pid": true,
	}

	// statFn stub: /usr/bin/nsenter always exists; only the listed ns files exist.
	statFn = func(path string) (os.FileInfo, error) {
		if path == "/usr/bin/nsenter" {
			return nil, nil
		}

		if presentNSFiles[path] {
			return nil, nil
		}

		return nil, os.ErrNotExist
	}

	// nsFlags that must always appear (from the present map above).
	expectedNSFlags := []string{"-m", "-n", "-p"}

	t.Run("timeNSJoinableFn returns true — -T present and ns flags included", func(t *testing.T) {
		timeNSJoinableFn = func() bool { return true }

		cmd, err := nsenterCommandWrapper(1000, 1000, []uint32{1000}, "/home/user", "/bin/sh")
		assert.NoError(t, err)

		// -T must be present.
		assert.Contains(t, cmd, "-T")

		// All namespace flags from statFn must be present.
		for _, flag := range expectedNSFlags {
			assert.Contains(t, cmd, flag)
		}

		// Flags for absent namespaces must NOT be present.
		for _, absent := range []string{"-u", "-i", "-C"} {
			assert.NotContains(t, cmd, absent)
		}

		// setpriv/nsenter prefix ordering: /bin/setpriv must come first,
		// /usr/bin/nsenter must follow.
		setprivIdx := indexOf(cmd, "/bin/setpriv")
		nsenterIdx := indexOf(cmd, "/usr/bin/nsenter")
		assert.NotEqual(t, -1, setprivIdx, "/bin/setpriv must be present")
		assert.NotEqual(t, -1, nsenterIdx, "/usr/bin/nsenter must be present")
		assert.Less(t, setprivIdx, nsenterIdx, "/bin/setpriv must precede /usr/bin/nsenter")

		// The user command must appear at the end.
		assert.Equal(t, "/bin/sh", cmd[len(cmd)-1])
	})

	t.Run("timeNSJoinableFn returns false — -T absent and ns flags still included", func(t *testing.T) {
		timeNSJoinableFn = func() bool { return false }

		cmd, err := nsenterCommandWrapper(1000, 1000, []uint32{1000}, "/home/user", "/bin/sh")
		assert.NoError(t, err)

		// -T must be absent.
		assert.NotContains(t, cmd, "-T")

		// All namespace flags from statFn must still be present.
		for _, flag := range expectedNSFlags {
			assert.Contains(t, cmd, flag)
		}

		// Flags for absent namespaces must NOT be present.
		for _, absent := range []string{"-u", "-i", "-C"} {
			assert.NotContains(t, cmd, absent)
		}

		// Prefix ordering must still be correct.
		setprivIdx := indexOf(cmd, "/bin/setpriv")
		nsenterIdx := indexOf(cmd, "/usr/bin/nsenter")
		assert.NotEqual(t, -1, setprivIdx, "/bin/setpriv must be present")
		assert.NotEqual(t, -1, nsenterIdx, "/usr/bin/nsenter must be present")
		assert.Less(t, setprivIdx, nsenterIdx, "/bin/setpriv must precede /usr/bin/nsenter")

		// The user command must appear at the end.
		assert.Equal(t, "/bin/sh", cmd[len(cmd)-1])
	})

	t.Run("statFn controls which ns flags appear — timeNS true", func(t *testing.T) {
		// Override statFn so only net is present.
		statFn = func(path string) (os.FileInfo, error) {
			if path == "/usr/bin/nsenter" || path == "/proc/1/ns/net" {
				return nil, nil
			}

			return nil, os.ErrNotExist
		}

		timeNSJoinableFn = func() bool { return true }

		cmd, err := nsenterCommandWrapper(1000, 1000, []uint32{1000}, "/home/user", "/bin/bash")
		assert.NoError(t, err)

		assert.Contains(t, cmd, "-n")
		assert.Contains(t, cmd, "-T")

		for _, absent := range []string{"-m", "-u", "-i", "-p", "-C"} {
			assert.NotContains(t, cmd, absent)
		}
	})

	t.Run("statFn controls which ns flags appear — timeNS false", func(t *testing.T) {
		// Override statFn so only mnt and uts are present.
		statFn = func(path string) (os.FileInfo, error) {
			if path == "/usr/bin/nsenter" || path == "/proc/1/ns/mnt" || path == "/proc/1/ns/uts" {
				return nil, nil
			}

			return nil, os.ErrNotExist
		}

		timeNSJoinableFn = func() bool { return false }

		cmd, err := nsenterCommandWrapper(1000, 1000, []uint32{1000}, "/home/user", "/bin/bash")
		assert.NoError(t, err)

		assert.Contains(t, cmd, "-m")
		assert.Contains(t, cmd, "-u")
		assert.NotContains(t, cmd, "-T")

		for _, absent := range []string{"-i", "-n", "-p", "-C"} {
			assert.NotContains(t, cmd, absent)
		}
	})
}

// indexOf returns the index of target in slice, or -1 if not found.
func indexOf(slice []string, target string) int {
	for i, s := range slice {
		if s == target {
			return i
		}
	}

	return -1
}

// TestProbeTimeNS verifies that nsenterCommandWrapper wires timeNSJoinableFn
// into nsenterArgs' joinTime argument correctly.
func TestProbeTimeNS(t *testing.T) {
	// Save and restore the real statFn and timeNSJoinableFn after each test.
	origStatFn := statFn
	origTimeNSJoinableFn := timeNSJoinableFn

	t.Cleanup(func() {
		statFn = origStatFn
		timeNSJoinableFn = origTimeNSJoinableFn
	})

	// Make statFn always return ErrNotExist so namespace probing skips all
	// entries — we only care about the time-namespace behaviour here.
	statFn = func(_ string) (os.FileInfo, error) {
		return nil, os.ErrNotExist
	}

	t.Run("timeNSJoinableFn returns true → -T present in command", func(t *testing.T) {
		timeNSJoinableFn = func() bool { return true }

		cmd, err := nsenterCommandWrapper(1000, 1000, nil, "/home/user", "/bin/sh")
		assert.NoError(t, err)
		assert.Contains(t, cmd, "-T")
	})

	t.Run("timeNSJoinableFn returns false → -T absent from command", func(t *testing.T) {
		timeNSJoinableFn = func() bool { return false }

		cmd, err := nsenterCommandWrapper(1000, 1000, nil, "/home/user", "/bin/sh")
		assert.NoError(t, err)
		assert.NotContains(t, cmd, "-T")
	})

	t.Run("timeNSJoinableFn transient false → -T absent from command", func(t *testing.T) {
		// Even a transient/error return of false must keep -T out.
		timeNSJoinableFn = func() bool { return false }

		cmd, err := nsenterCommandWrapper(1000, 1000, nil, "/home/user", "/bin/sh")
		assert.NoError(t, err)
		assert.NotContains(t, cmd, "-T")
	})
}

func TestNsenterArgs(t *testing.T) {
	t.Run("joinTime=false excludes -T", func(t *testing.T) {
		present := map[string]string{
			"mnt": "-m",
			"net": "-n",
		}
		args := nsenterArgs(present, false)

		assert.NotContains(t, args, "-T")
		assert.Contains(t, args, "-m")
		assert.Contains(t, args, "-n")
	})

	t.Run("joinTime=true includes -T", func(t *testing.T) {
		present := map[string]string{
			"mnt": "-m",
		}
		args := nsenterArgs(present, true)

		assert.Contains(t, args, "-T")
		assert.Contains(t, args, "-m")
	})

	t.Run("all flags in present are forwarded unchanged", func(t *testing.T) {
		present := map[string]string{
			"mnt":    "-m",
			"uts":    "-u",
			"ipc":    "-i",
			"net":    "-n",
			"pid":    "-p",
			"cgroup": "-C",
		}
		args := nsenterArgs(present, false)

		got := make([]string, len(args))
		copy(got, args)
		sort.Strings(got)

		expected := []string{"-C", "-i", "-m", "-n", "-p", "-u"}
		assert.Equal(t, expected, got)
	})

	t.Run("empty present with joinTime=true returns only -T", func(t *testing.T) {
		present := map[string]string{}
		args := nsenterArgs(present, true)

		assert.Equal(t, []string{"-T"}, args)
	})

	t.Run("empty present with joinTime=false returns empty slice", func(t *testing.T) {
		present := map[string]string{}
		args := nsenterArgs(present, false)

		assert.Empty(t, args)
	})
}

// TestTimeNSMemoizer verifies the tri-state memoizer behaviour injected via
// newTimeNSMemoizer.  Tests do NOT call t.Parallel() because they mutate
// package-level variables.
//
//nolint:paralleltest
func TestTimeNSMemoizer(t *testing.T) {
	t.Run("definitive result is cached — probe called exactly once", func(t *testing.T) {
		var calls atomic.Int64

		// Probe returns (true, true) → definitive success, must be cached.
		probe := func() (bool, bool) {
			calls.Add(1)

			return true, true
		}

		fn := newTimeNSMemoizer(probe)

		const n = 10

		for i := range n {
			_ = i
			got := fn()
			assert.True(t, got)
		}

		assert.Equal(t, int64(1), calls.Load(), "probe must be called exactly once for a definitive result")
	})

	t.Run("transient result is NOT cached — probe called each time", func(t *testing.T) {
		var calls atomic.Int64

		// Probe returns (false, false) → transient failure, must NOT be cached.
		probe := func() (bool, bool) {
			calls.Add(1)

			return false, false
		}

		fn := newTimeNSMemoizer(probe)

		const n = 5

		for i := range n {
			_ = i
			got := fn()
			assert.False(t, got)
		}

		assert.Equal(t, int64(n), calls.Load(), "probe must be called every time for a transient result")
	})

	t.Run("misconfiguration result is NOT cached — probe called each time", func(t *testing.T) {
		var calls atomic.Int64

		// Probe returns (false, false) → misconfiguration (e.g. nsenter not
		// found), must NOT be cached so the next call retries.
		probe := func() (bool, bool) {
			calls.Add(1)

			return false, false
		}

		fn := newTimeNSMemoizer(probe)

		const n = 4

		for i := range n {
			_ = i
			got := fn()
			assert.False(t, got)
		}

		assert.Equal(t, int64(n), calls.Load(), "probe must be called every time for a misconfiguration result")
	})

	t.Run("definitive false is cached — probe called exactly once", func(t *testing.T) {
		var calls atomic.Int64

		// Probe returns (false, true) → definitive denial, must be cached.
		probe := func() (bool, bool) {
			calls.Add(1)

			return false, true
		}

		fn := newTimeNSMemoizer(probe)

		const n = 6

		for i := range n {
			_ = i
			got := fn()
			assert.False(t, got)
		}

		assert.Equal(t, int64(1), calls.Load(), "probe must be called exactly once for a definitive-false result")
	})
}
