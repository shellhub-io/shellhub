//go:build !docker
// +build !docker

package command

import (
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"syscall"

	"github.com/shellhub-io/shellhub/pkg/agent/pkg/osauth"
	log "github.com/sirupsen/logrus"
)

func NewCmd(u *osauth.User, shell, term, host string, envs []string, command ...string) *exec.Cmd {
	user, _ := user.Lookup(u.Username)
	userGroups, _ := user.GroupIds()

	// Supplementary groups for the user
	groups := make([]uint32, 0)
	for _, sgid := range userGroups {
		igid, _ := strconv.ParseUint(sgid, 10, 32)
		groups = append(groups, uint32(igid)) //nolint:gosec // The value of igid fits inside a uint32.
	}
	if len(groups) == 0 {
		groups = append(groups, u.GID)
	}

	// NOTE: The `exec.Command` only resolves, as documentation says, the first item on the command slice, what is, in
	// this case, the shell (e.g. /bin/bash). So we need to resolve the target command path by ourselves.
	// If we don't do that, the shell will try to find the command in its own PATH, what may lead to a "command not
	// found" error if the PATH is not correctly set.
	target, err := exec.LookPath(command[2])
	if err != nil {
		log.WithError(err).
			WithFields(log.Fields{
				"command": command[2],
				"path":    os.Getenv("PATH"),
			}).
			Warn("failed to resolve command path, using the command as is")

		target = command[2]
	}

	// NOTE: After the resolution, we set the command[2] to the resolved path.
	command[2] = target

	// TODO: Remove this debug line before merge.
	log.Println(os.Getenv("PATH"))

	cmd := exec.Command(command[0], command[1:]...) //nolint:gosec
	cmd.Env = []string{
		"TERM=" + term,
		"HOME=" + u.HomeDir,
		"SHELL=" + shell,
		"SHELLHUB_HOST=" + host,
	}
	cmd.Env = append(cmd.Env, envs...)

	if _, err := os.Stat(u.HomeDir); err != nil {
		log.WithError(err).WithField("dir", u.HomeDir).Warn("setting user's home directory to /")

		cmd.Dir = "/"
	} else {
		cmd.Dir = u.HomeDir
	}

	if os.Geteuid() == 0 {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
		cmd.SysProcAttr.Credential = &syscall.Credential{Uid: u.UID, Gid: u.GID, Groups: groups}
	}

	return cmd
}

// SFTPServerCommand creates the command used by agent to start the SFTP server used in a SFTP connection.
func SFTPServerCommand() *exec.Cmd {
	return exec.Command("/proc/self/exe", []string{"sftp", string(SFTPServerModeNative)}...) //nolint:gosec
}
