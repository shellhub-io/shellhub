package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"

	"github.com/pkg/sftp"
	"github.com/shellhub-io/shellhub/agent/server/modes/host/command"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const sftpServerChildEnv = "SHELLHUB_TEST_SFTP_SERVER_CHILD"

func TestMain(m *testing.M) {
	if os.Getenv(sftpServerChildEnv) == "1" {
		NewSFTPServer(command.SFTPServerModeNative)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func TestReportServeError(t *testing.T) {
	cases := []struct {
		name     string
		err      error
		expected string
	}{
		{name: "clean end reports nothing", err: nil, expected: ""},
		{name: "client EOF reports nothing", err: io.EOF, expected: ""},
		{name: "wrapped client EOF reports nothing", err: fmt.Errorf("serve: %w", io.EOF), expected: ""},
		{name: "real error is reported", err: errors.New("connection reset"), expected: "connection reset\n"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var stderr bytes.Buffer

			reportServeError(&stderr, tc.err)

			assert.Equal(t, tc.expected, stderr.String())
		})
	}
}

func TestSFTPCredentialsFromEnv(t *testing.T) {
	cases := []struct {
		name        string
		env         map[string]string
		expected    sftpCredentials
		requireErrs require.ErrorAssertionFunc
	}{
		{
			name:        "user with supplementary groups",
			env:         map[string]string{"HOME": "/home/alice", "UID": "1000", "GID": "1000", "SUPPLEMENTARY_GIDS": "4,27"},
			expected:    sftpCredentials{home: "/home/alice", uid: 1000, gid: 1000, groups: []int{4, 27}},
			requireErrs: require.NoError,
		},
		{
			name:        "user without supplementary groups",
			env:         map[string]string{"HOME": "/home/alice", "UID": "1000", "GID": "1000", "SUPPLEMENTARY_GIDS": ""},
			expected:    sftpCredentials{home: "/home/alice", uid: 1000, gid: 1000, groups: []int{}},
			requireErrs: require.NoError,
		},
		{
			name:        "missing supplementary gids drop every supplementary group",
			env:         map[string]string{"HOME": "/home/alice", "UID": "1000", "GID": "1000"},
			expected:    sftpCredentials{home: "/home/alice", uid: 1000, gid: 1000, groups: []int{}},
			requireErrs: require.NoError,
		},
		{
			name:        "malformed supplementary gids",
			env:         map[string]string{"HOME": "/home/alice", "UID": "1000", "GID": "1000", "SUPPLEMENTARY_GIDS": "wheel"},
			requireErrs: require.Error,
		},
		{
			name:        "malformed uid",
			env:         map[string]string{"HOME": "/home/alice", "UID": "alice", "GID": "1000", "SUPPLEMENTARY_GIDS": ""},
			requireErrs: require.Error,
		},
		{
			name:        "malformed gid",
			env:         map[string]string{"HOME": "/home/alice", "UID": "1000", "GID": "alice", "SUPPLEMENTARY_GIDS": ""},
			requireErrs: require.Error,
		},
		{
			name:        "missing home",
			env:         map[string]string{"UID": "1000", "GID": "1000", "SUPPLEMENTARY_GIDS": ""},
			requireErrs: require.Error,
		},
		{
			name:        "missing uid",
			env:         map[string]string{"HOME": "/home/alice", "GID": "1000", "SUPPLEMENTARY_GIDS": ""},
			requireErrs: require.Error,
		},
		{
			name:        "missing gid",
			env:         map[string]string{"HOME": "/home/alice", "UID": "1000", "SUPPLEMENTARY_GIDS": ""},
			requireErrs: require.Error,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			creds, err := sftpCredentialsFromEnv(func(key string) (string, bool) {
				value, ok := tc.env[key]

				return value, ok
			})

			tc.requireErrs(t, err)
			if err == nil {
				assert.Equal(t, tc.expected, creds)
			}
		})
	}
}

type recordedPrivilegeSyscalls struct {
	euid  int
	fail  string
	calls []string
}

func (r *recordedPrivilegeSyscalls) syscalls() privilegeSyscalls {
	record := func(name string) error {
		r.calls = append(r.calls, name)
		if name == r.fail {
			return syscall.EPERM
		}

		return nil
	}

	return privilegeSyscalls{
		geteuid:   func() int { return r.euid },
		setgroups: func(groups []int) error { return record(fmt.Sprintf("setgroups%v", groups)) },
		setgid:    func(gid int) error { return record(fmt.Sprintf("setgid(%d)", gid)) },
		setuid:    func(uid int) error { return record(fmt.Sprintf("setuid(%d)", uid)) },
	}
}

func TestPrivilegeSyscallsSwitchTo(t *testing.T) {
	creds := sftpCredentials{home: "/home/alice", uid: 1000, gid: 1001, groups: []int{4, 27}}

	cases := []struct {
		name        string
		euid        int
		creds       sftpCredentials
		fail        string
		calls       []string
		requireErrs require.ErrorAssertionFunc
	}{
		{
			name:        "root replaces the groups before the gid and the uid",
			euid:        0,
			creds:       creds,
			calls:       []string{"setgroups[4 27]", "setgid(1001)", "setuid(1000)"},
			requireErrs: require.NoError,
		},
		{
			name:        "root clears the groups of a user without supplementary groups",
			euid:        0,
			creds:       sftpCredentials{home: "/home/alice", uid: 1000, gid: 1001, groups: []int{}},
			calls:       []string{"setgroups[]", "setgid(1001)", "setuid(1000)"},
			requireErrs: require.NoError,
		},
		{
			name:        "non-root agent switches nothing",
			euid:        1000,
			creds:       creds,
			calls:       nil,
			requireErrs: require.NoError,
		},
		{
			name:        "refused setgroups stops before the gid and the uid",
			euid:        0,
			creds:       creds,
			fail:        "setgroups[4 27]",
			calls:       []string{"setgroups[4 27]"},
			requireErrs: require.Error,
		},
		{
			name:        "refused setgid stops before the uid",
			euid:        0,
			creds:       creds,
			fail:        "setgid(1001)",
			calls:       []string{"setgroups[4 27]", "setgid(1001)"},
			requireErrs: require.Error,
		},
		{
			name:        "refused setuid is reported",
			euid:        0,
			creds:       creds,
			fail:        "setuid(1000)",
			calls:       []string{"setgroups[4 27]", "setgid(1001)", "setuid(1000)"},
			requireErrs: require.Error,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			recorder := &recordedPrivilegeSyscalls{euid: tc.euid, fail: tc.fail}

			tc.requireErrs(t, recorder.syscalls().switchTo(tc.creds))
			assert.Equal(t, tc.calls, recorder.calls)
		})
	}
}

func TestSFTPServer_RunsWithTheUsersGroupsOnly(t *testing.T) {
	if os.Geteuid() != 0 {
		t.Skip("switching to another user's credentials needs root")
	}

	const agentOnlyGroup = 4242

	original, err := syscall.Getgroups()
	require.NoError(t, err)
	require.NoError(t, syscall.Setgroups(append(original, agentOnlyGroup)))
	t.Cleanup(func() {
		require.NoError(t, syscall.Setgroups(original))
	})

	dir := t.TempDir()
	for _, traversed := range []string{filepath.Dir(dir), dir} {
		require.NoError(t, os.Chmod(traversed, 0o755)) //nolint:gosec // the SFTP user must traverse the directory
	}

	secret := filepath.Join(dir, "group-only-secret")
	require.NoError(t, os.WriteFile(secret, []byte("group-only-secret"), 0o600))
	require.NoError(t, os.Chown(secret, 0, agentOnlyGroup))
	require.NoError(t, os.Chmod(secret, 0o640)) //nolint:gosec // group-readable is the permission under test

	cases := []struct {
		name         string
		groups       string
		statusGroups string
		readable     bool
	}{
		{name: "user outside the agent's groups", groups: "", statusGroups: "", readable: false},
		{name: "user in the file's group", groups: strconv.Itoa(agentOnlyGroup), statusGroups: strconv.Itoa(agentOnlyGroup), readable: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			client := startSFTPServerAs(t, dir, 65534, 65534, tc.groups)

			assert.Equal(t, tc.statusGroups, sftpServerGroups(t, client))

			file, err := client.Open(secret)
			if !tc.readable {
				require.ErrorIs(t, err, os.ErrPermission)

				return
			}

			require.NoError(t, err)
			defer func() { assert.NoError(t, file.Close()) }()

			content, err := io.ReadAll(file)
			require.NoError(t, err)
			assert.Equal(t, "group-only-secret", string(content))
		})
	}
}

func startSFTPServerAs(t *testing.T, home string, uid, gid int, groups string) *sftp.Client {
	t.Helper()

	cmd := exec.CommandContext(t.Context(), os.Args[0]) //nolint:gosec // re-executes the test binary as the SFTP server
	cmd.Env = []string{
		sftpServerChildEnv + "=1",
		"HOME=" + home,
		fmt.Sprintf("UID=%d", uid),
		fmt.Sprintf("GID=%d", gid),
		"SUPPLEMENTARY_GIDS=" + groups,
	}

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	input, err := cmd.StdinPipe()
	require.NoError(t, err)

	output, err := cmd.StdoutPipe()
	require.NoError(t, err)

	require.NoError(t, cmd.Start())

	client, err := sftp.NewClientPipe(output, input)
	require.NoError(t, err, stderr.String())

	t.Cleanup(func() {
		_ = client.Close()
		_ = cmd.Wait()
	})

	return client
}

func sftpServerGroups(t *testing.T, client *sftp.Client) string {
	t.Helper()

	file, err := client.Open("/proc/self/status")
	require.NoError(t, err)
	defer func() { assert.NoError(t, file.Close()) }()

	status, err := io.ReadAll(file)
	require.NoError(t, err)

	for line := range strings.SplitSeq(string(status), "\n") {
		if groups, ok := strings.CutPrefix(line, "Groups:"); ok {
			return strings.TrimSpace(groups)
		}
	}

	require.FailNow(t, "no Groups line in /proc/self/status")

	return ""
}
