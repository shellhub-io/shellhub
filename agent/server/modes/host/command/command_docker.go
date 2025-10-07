//go:build docker
// +build docker

package command

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"

	"github.com/shellhub-io/shellhub/agent/pkg/osauth"
)

func NewCmd(u *osauth.User, shell, term, host string, envs []string, command ...string) *exec.Cmd {
	nscommand, _ := nsenterCommandWrapper(u.UID, u.GID, u.HomeDir, command...)

	cmd := exec.Command(nscommand[0], nscommand[1:]...) //nolint:gosec
	// TODO: There are other environment variables we could set like SSH_CONNECTION, SSH_TTY, SSH_ORIGINAL_COMMAND, etc.
	// We need to check which ones are relevant and set them accordingly.
	// https://en.wikibooks.org/wiki/OpenSSH/Client_Applications
	cmd.Env = []string{
		"TERM=" + term,
		"HOME=" + u.HomeDir,
		"SHELL=" + shell,
		"USER=" + u.Username,
		"LOGNAME=" + u.Username,
		"SHELLHUB_HOST=" + host,
		// NOTE: We need to set the SSH_CLIENT because some applications (like bash) check for it to enable some
		// features or load some files (like .bashrc). Currently, we don't have this information, so we set a fake one.
		// TODO: Set the real SSH_CLIENT value.
		// Format: "<ip> <source-port> <destination-port>"
		// https://en.wikibooks.org/wiki/OpenSSH/Client_Applications
		"SSH_CLIENT=127.0.0.1 0 0",
	}
	cmd.Env = append(cmd.Env, envs...)

	return cmd
}

func getWrappedCommand(nsArgs []string, uid, gid uint32, home string) []string {
	setPrivCmd := []string{
		"/usr/bin/setpriv",
		"--init-groups",
		"--ruid",
		strconv.Itoa(int(uid)),
		"--regid",
		strconv.Itoa(int(gid)),
	}

	nsenterCmd := append([]string{
		"/usr/bin/nsenter",
		"-t",
		"1",
	}, nsArgs...)

	nsenterCmd = append(nsenterCmd,
		[]string{
			"-S",
			strconv.Itoa(int(uid)),
			fmt.Sprintf("--wdns=%s", home),
		}...,
	)

	return append(setPrivCmd, nsenterCmd...)
}

func nsenterCommandWrapper(uid, gid uint32, home string, command ...string) ([]string, error) {
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

	return append(getWrappedCommand(args, uid, gid, home), command...), nil
}

// SFTPServerCommand creates the command used by agent to start the SFTP server used in a SFTP connection.
func SFTPServerCommand() *exec.Cmd {
	return exec.Command("/proc/self/exe", []string{"sftp", string(SFTPServerModeDocker)}...) //nolint:gosec
}
