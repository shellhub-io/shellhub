//go:build !docker

package command

import (
	"errors"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/shellhub-io/shellhub/agent/pkg/osauth"
	log "github.com/sirupsen/logrus"
)

var geteuidFn = os.Geteuid

var readSetgroupsPolicyFn = func() ([]byte, error) {
	return os.ReadFile("/proc/self/setgroups")
}

func setgroupsDenied() bool {
	data, err := readSetgroupsPolicyFn()
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			log.WithError(err).Warn("failed to read /proc/self/setgroups; assuming setgroups is allowed")
		}

		return false
	}

	return strings.TrimSpace(string(data)) == "deny"
}

// CheckCredentialSwitch reports whether the process can switch credentials via
// setgroups(2).  It is a pre-flight check that must be called before attempting
// to execute a command as a different user.
//
// Short-circuit: when the effective UID is not root (euid != 0), credential
// switching is a no-op and the check always succeeds (nil).
//
// When euid == 0, the kernel may still forbid setgroups inside an unprivileged
// user namespace (Linux ≥ 3.19).  In that case the function returns a sentinel
// error whose message contains "setgroups denied in unprivileged user namespace".
func CheckCredentialSwitch() error {
	if geteuidFn() != 0 {
		return nil
	}

	if setgroupsDenied() {
		return errors.New("setgroups denied in unprivileged user namespace")
	}

	return nil
}

// NewCmd builds the process a session runs as u, with the environment and supplementary
// groups the account is entitled to. A group lookup failure leaves the set empty rather
// than failing the session.
func NewCmd(u *osauth.User, shell, term, host string, envs []string, command ...string) *exec.Cmd {
	groups, err := osauth.ListGroups(u.Username)
	if err != nil {
		groups = []uint32{}
	}

	cmd := exec.Command(command[0], command[1:]...) //nolint:noctx,gosec
	cmd.Env = []string{
		"TERM=" + term,
		"HOME=" + u.HomeDir,
		"SHELL=" + shell,
		"USER=" + u.Username,
		"LOGNAME=" + u.Username,
		"SHELLHUB_HOST=" + host,
		"SSH_CLIENT=127.0.0.1 0 0",
	}
	cmd.Env = append(cmd.Env, envs...)

	if _, err := os.Stat(u.HomeDir); err != nil {
		log.WithError(err).WithField("dir", u.HomeDir).Warn("setting user's home directory to /")

		cmd.Dir = "/"
	} else {
		cmd.Dir = u.HomeDir
	}

	if geteuidFn() == 0 {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
		cmd.SysProcAttr.Credential = &syscall.Credential{Uid: u.UID, Gid: u.GID, Groups: groups}
	}

	return cmd
}

// SFTPServerCommand creates the command used by agent to start the SFTP server used in a SFTP connection.
func SFTPServerCommand() *exec.Cmd {
	return exec.Command("/proc/self/exe", []string{"sftp", string(SFTPServerModeNative)}...) //nolint:noctx,gosec
}
