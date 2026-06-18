//go:build !docker
// +build !docker

package command

import (
	"errors"
	"os"
	"sync/atomic"
	"testing"

	"github.com/shellhub-io/shellhub/agent/pkg/osauth"
	"github.com/stretchr/testify/assert"
)

// Compile-time assertion: CheckCredentialSwitch must be defined under the !docker tag.
var _ = CheckCredentialSwitch

// TestSetgroupsDenied verifies the setgroupsDenied helper across all documented
// scenarios.  Tests do NOT call t.Parallel() because they mutate package-level
// seam variables.
//
//nolint:paralleltest
func TestSetgroupsDenied(t *testing.T) {
	origGeteuidFn := geteuidFn
	origReadSetgroupsPolicyFn := readSetgroupsPolicyFn

	t.Cleanup(func() {
		geteuidFn = origGeteuidFn
		readSetgroupsPolicyFn = origReadSetgroupsPolicyFn
	})

	t.Run("'deny' returns true", func(t *testing.T) {
		readSetgroupsPolicyFn = func() ([]byte, error) {
			return []byte("deny"), nil
		}

		assert.True(t, setgroupsDenied())
	})

	t.Run("'deny\\n' returns true", func(t *testing.T) {
		readSetgroupsPolicyFn = func() ([]byte, error) {
			return []byte("deny\n"), nil
		}

		assert.True(t, setgroupsDenied())
	})

	t.Run("' deny \\n' returns true", func(t *testing.T) {
		readSetgroupsPolicyFn = func() ([]byte, error) {
			return []byte(" deny \n"), nil
		}

		assert.True(t, setgroupsDenied())
	})

	t.Run("'allow' returns false", func(t *testing.T) {
		readSetgroupsPolicyFn = func() ([]byte, error) {
			return []byte("allow"), nil
		}

		assert.False(t, setgroupsDenied())
	})

	t.Run("ErrNotExist returns false without warning", func(t *testing.T) {
		readSetgroupsPolicyFn = func() ([]byte, error) {
			return nil, os.ErrNotExist
		}

		// No warning expected — just false.
		assert.False(t, setgroupsDenied())
	})

	t.Run("other read error returns false with warning", func(t *testing.T) {
		readSetgroupsPolicyFn = func() ([]byte, error) {
			return nil, errors.New("permission denied")
		}

		// Should return false (and emit a warning via log, but we do not
		// capture log output — the return value is what matters here).
		assert.False(t, setgroupsDenied())
	})
}

// TestCheckCredentialSwitch verifies CheckCredentialSwitch across its four
// documented scenarios.  Tests do NOT call t.Parallel() because they mutate
// package-level seam variables.
//
//nolint:paralleltest
func TestCheckCredentialSwitch(t *testing.T) {
	origGeteuidFn := geteuidFn
	origReadSetgroupsPolicyFn := readSetgroupsPolicyFn

	t.Cleanup(func() {
		geteuidFn = origGeteuidFn
		readSetgroupsPolicyFn = origReadSetgroupsPolicyFn
	})

	t.Run("euid!=0 returns nil without reading setgroups policy", func(t *testing.T) {
		var calls atomic.Int64

		geteuidFn = func() int { return 1000 }
		readSetgroupsPolicyFn = func() ([]byte, error) {
			calls.Add(1)

			return []byte("deny"), nil
		}

		err := CheckCredentialSwitch()

		assert.NoError(t, err)
		assert.Equal(t, int64(0), calls.Load(), "readSetgroupsPolicyFn must never be called when euid!=0")
	})

	t.Run("euid!=0 with deny policy still returns nil (stub never called)", func(t *testing.T) {
		var calls atomic.Int64

		geteuidFn = func() int { return 500 }
		readSetgroupsPolicyFn = func() ([]byte, error) {
			calls.Add(1)

			return []byte("deny"), nil
		}

		err := CheckCredentialSwitch()

		assert.NoError(t, err)
		assert.Equal(t, int64(0), calls.Load(), "readSetgroupsPolicyFn must never be called when euid!=0")
	})

	t.Run("euid==0 and setgroups deny returns error containing 'setgroups denied in unprivileged user namespace'", func(t *testing.T) {
		geteuidFn = func() int { return 0 }
		readSetgroupsPolicyFn = func() ([]byte, error) {
			return []byte("deny"), nil
		}

		err := CheckCredentialSwitch()

		assert.Error(t, err)
		assert.ErrorContains(t, err, "setgroups denied in unprivileged user namespace")
	})

	t.Run("euid==0 and setgroups allow returns nil", func(t *testing.T) {
		geteuidFn = func() int { return 0 }
		readSetgroupsPolicyFn = func() ([]byte, error) {
			return []byte("allow"), nil
		}

		err := CheckCredentialSwitch()

		assert.NoError(t, err)
	})
}

func containsEnv(envs []string, entry string) bool {
	for _, v := range envs {
		if v == entry {
			return true
		}
	}

	return false
}

func TestNewCmdNativeEnv(t *testing.T) {
	tests := []struct {
		name string
		user *osauth.User
		want []string
	}{
		{
			name: "USER and LOGNAME are set from username",
			user: &osauth.User{
				Username: "bob",
				HomeDir:  "/",
				Shell:    "/bin/sh",
			},
			want: []string{
				"USER=bob",
				"LOGNAME=bob",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewCmd(tt.user, "/bin/sh", "xterm", "myhost", nil, "/bin/sh")

			for _, entry := range tt.want {
				assert.True(t, containsEnv(cmd.Env, entry))
			}
		})
	}
}
