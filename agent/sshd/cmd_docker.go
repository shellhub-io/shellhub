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

func nsenterCommandWrapper(uid, gid uint32, home string, command ...string) ([]string, error) {
	wrappedCommand := []string{}

	if _, err := os.Stat("/usr/bin/nsenter"); err == nil {
		wrappedCommand = append([]string{
			"/usr/bin/setpriv",
			"--init-groups",
			"--ruid",
			strconv.Itoa(int(uid)),
			"--regid",
			strconv.Itoa(int(gid)),
			"/usr/bin/nsenter",
			"-t", "1",
			"-a",
			"-S",
			strconv.Itoa(int(uid)),
			fmt.Sprintf("--wd=%s", home),
		}, wrappedCommand...)
	} else if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	wrappedCommand = append(wrappedCommand, command...)

	return wrappedCommand, nil
}
