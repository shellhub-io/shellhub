package sysinfo_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/shellhub-io/shellhub/agent/pkg/sysinfo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetOSRelease(t *testing.T) {
	cases := []struct {
		description string
		etc         *string
		usrLib      *string
		expected    *sysinfo.OSRelease
	}{
		{
			description: "takes ID and PRETTY_NAME",
			etc:         new("ID=ubuntu\nNAME=\"Ubuntu\"\nPRETTY_NAME=\"Ubuntu 24.04 LTS\"\n"),
			expected:    &sysinfo.OSRelease{ID: "ubuntu", Name: "Ubuntu 24.04 LTS"},
		},
		{
			description: "falls back to NAME when PRETTY_NAME is unset",
			etc:         new("ID=alpine\nNAME=\"Alpine Linux\"\n"),
			expected:    &sysinfo.OSRelease{ID: "alpine", Name: "Alpine Linux"},
		},
		{
			description: "falls back to NAME when PRETTY_NAME is empty",
			etc:         new("ID=alpine\nNAME=\"Alpine Linux\"\nPRETTY_NAME=\"\"\n"),
			expected:    &sysinfo.OSRelease{ID: "alpine", Name: "Alpine Linux"},
		},
		{
			description: "falls back to Linux when no name key is set",
			etc:         new("ID=custom\n"),
			expected:    &sysinfo.OSRelease{ID: "custom", Name: "Linux"},
		},
		{
			description: "falls back to linux when ID is unset",
			etc:         new("PRETTY_NAME=\"Custom OS\"\n"),
			expected:    &sysinfo.OSRelease{ID: "linux", Name: "Custom OS"},
		},
		{
			description: "falls back to linux when ID is empty",
			etc:         new("ID=\"\"\nPRETTY_NAME=\"Custom OS\"\n"),
			expected:    &sysinfo.OSRelease{ID: "linux", Name: "Custom OS"},
		},
		{
			description: "reads /usr/lib/os-release when /etc/os-release does not exist",
			usrLib:      new("ID=fedora\nPRETTY_NAME=\"Fedora Linux 40\"\n"),
			expected:    &sysinfo.OSRelease{ID: "fedora", Name: "Fedora Linux 40"},
		},
		{
			description: "prefers /etc/os-release over /usr/lib/os-release",
			etc:         new("ID=debian\nPRETTY_NAME=\"Debian GNU/Linux 12\"\n"),
			usrLib:      new("ID=fedora\nPRETTY_NAME=\"Fedora Linux 40\"\n"),
			expected:    &sysinfo.OSRelease{ID: "debian", Name: "Debian GNU/Linux 12"},
		},
		{
			description: "does not fill keys missing from /etc/os-release from /usr/lib/os-release",
			etc:         new("VERSION_ID=12\n"),
			usrLib:      new("ID=fedora\nPRETTY_NAME=\"Fedora Linux 40\"\n"),
			expected:    &sysinfo.OSRelease{ID: "linux", Name: "Linux"},
		},
		{
			description: "returns the defaults when no os-release file exists",
			expected:    &sysinfo.OSRelease{ID: "linux", Name: "Linux"},
		},
		{
			description: "reads quoted and unquoted values and skips comments",
			etc:         new("# comment\n\nID=arch\nPRETTY_NAME='Arch Linux'\n"),
			expected:    &sysinfo.OSRelease{ID: "arch", Name: "Arch Linux"},
		},
		{
			description: "skips a line whose value does not parse",
			etc:         new("ID=debian\nNAME=\"Debian GNU/Linux\"\nPRETTY_NAME=\"Debian GNU/Linux 12\n"),
			expected:    &sysinfo.OSRelease{ID: "debian", Name: "Debian GNU/Linux"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			root := t.TempDir()
			writeOSRelease(t, filepath.Join(root, "etc", "os-release"), tc.etc)
			writeOSRelease(t, filepath.Join(root, "usr", "lib", "os-release"), tc.usrLib)

			previous := sysinfo.OSReleaseRoot
			sysinfo.OSReleaseRoot = root
			t.Cleanup(func() { sysinfo.OSReleaseRoot = previous })

			got, err := sysinfo.GetOSRelease()
			require.NoError(t, err)
			assert.Equal(t, tc.expected, got)
		})
	}
}

func TestGetOSReleaseUnreadableFile(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, "etc", "os-release"), 0o750))

	previous := sysinfo.OSReleaseRoot
	sysinfo.OSReleaseRoot = root
	t.Cleanup(func() { sysinfo.OSReleaseRoot = previous })

	got, err := sysinfo.GetOSRelease()
	require.Error(t, err)
	assert.Equal(t, &sysinfo.OSRelease{ID: "linux", Name: "Linux"}, got)
}

func writeOSRelease(t *testing.T, path string, content *string) {
	t.Helper()

	if content == nil {
		return
	}

	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o750))
	require.NoError(t, os.WriteFile(path, []byte(*content), 0o600))
}
