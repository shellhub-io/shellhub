// +build docker

package sshd

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"

	"github.com/shellhub-io/shellhub/agent/pkg/osauth"
)

func newCmd(u *osauth.User, shell, term, host string, command ...string) *exec.Cmd {
	nscommand, _ := nsenterCommandWrapper(u.UID, u.GID, fmt.Sprintf("/host/%s", u.HomeDir), command...)

	cmd := exec.Command(nscommand[0], nscommand[1:]...)
	cmd.Env = []string{
		"TERM=" + term,
		"HOME=" + u.HomeDir,
		"SHELL=" + shell,
		"USER=" + u.Username,
		"LOGNAME=" + u.Username,
		"SHELLHUB_HOST=" + host,
	}

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
			fmt.Sprintf("--wd=%s", home),
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
