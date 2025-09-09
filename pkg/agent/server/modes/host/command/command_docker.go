//go:build docker
// +build docker

package command

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/shellhub-io/shellhub/pkg/agent/pkg/osauth"
)

func NewCmd(u *osauth.User, shell, term, host string, envs []string, command ...string) *exec.Cmd {
	nscommand, _ := nsenterCommandWrapper(u.UID, u.GID, u.HomeDir, u.Username, command...)

	cmd := exec.Command(nscommand[0], nscommand[1:]...) //nolint:gosec
	cmd.Env = []string{
		"TERM=" + term,
		"HOME=" + u.HomeDir,
		"SHELL=" + shell,
		"USER=" + u.Username,
		"LOGNAME=" + u.Username,
		"SHELLHUB_HOST=" + host,
	}
	cmd.Env = append(cmd.Env, envs...)

	return cmd
}

func getWrappedCommand(nsArgs []string, uid, gid uint32, home string, username string) []string {
	nsenterCmd := append([]string{
		"/usr/bin/nsenter",
		"-t",
		"1",
	}, nsArgs...)

	nsenterCmd = append(nsenterCmd,
		[]string{
			fmt.Sprintf("--wdns=%s", home),
		}...,
	)

	suCmd := []string{
		"su",
		"-",
		username,
		"-c",
	}

	return append(nsenterCmd, suCmd...)
}

func nsenterCommandWrapper(uid, gid uint32, home string, username string, command ...string) ([]string, error) {
	if _, err := os.Stat("/usr/bin/nsenter"); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	paths := map[string]string{
		"mnt":    "-m",
		"uts":    "-u",
		"ipc":    "-i",
		"net":    "-n",
		"pid":    "-p",
		"cgroup": "-C",
		"time":   "-T",
	}

	args := []string{}
	for path, params := range paths {
		if _, err := os.Stat(fmt.Sprintf("/proc/1/ns/%s", path)); err != nil {
			continue
		}

		args = append(args, params)
	}

	wrappedCmd := getWrappedCommand(args, uid, gid, home, username)

	// Join the command arguments into a single string for su -c
	fullCommand := ""
	for i, cmd := range command {
		if i > 0 {
			fullCommand += " "
		}
		fullCommand += cmd
	}

	return append(wrappedCmd, fullCommand), nil
}

// SFTPServerCommand creates the command used by agent to start the SFTP server used in a SFTP connection.
func SFTPServerCommand() *exec.Cmd {
	return exec.Command("/proc/self/exe", []string{"sftp", string(SFTPServerModeDocker)}...)
}
