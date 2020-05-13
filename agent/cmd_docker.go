// +build docker

package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"

	"github.com/shellhub-io/shellhub/agent/internal/osauth"
)

func newCmd(u *osauth.User, shell, term, host string, command ...string) *exec.Cmd {
	uid, _ := strconv.Atoi(u.UID)
	gid, _ := strconv.Atoi(u.GID)

	nscommand, _ := nsenterCommandWrapper(uid, gid, fmt.Sprintf("/host/%s", u.HomeDir), command...)

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

func nsenterCommandWrapper(uid, gid int, home string, command ...string) ([]string, error) {
	wrappedCommand := []string{}

	if _, err := os.Stat("/usr/bin/nsenter"); err == nil {
		wrappedCommand = append([]string{
			"/usr/bin/setpriv",
			"--init-groups",
			"--ruid",
			strconv.Itoa(uid),
			"--regid",
			strconv.Itoa(gid),
			"/usr/bin/nsenter",
			"-t", "1",
			"-a",
			"-S",
			strconv.Itoa(uid),
			fmt.Sprintf("--wd=%s", home),
		}, wrappedCommand...)
	} else if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	wrappedCommand = append(wrappedCommand, command...)

	return wrappedCommand, nil
}
