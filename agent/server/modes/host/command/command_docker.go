//go:build docker

package command

import (
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/shellhub-io/shellhub/agent/pkg/osauth"
)

var statFn = os.Stat

func nsenterArgs(present map[string]string) []string {
	args := []string{}

	for _, flag := range present {
		args = append(args, flag)
	}

	return args
}

// CheckCredentialSwitch is a no-op in Docker mode: the agent relies on
// nsenter+setpriv for credential switching, so this check is not applicable.
func CheckCredentialSwitch() error {
	return nil
}

func NewCmd(u *osauth.User, shell, term, host string, envs []string, command ...string) *exec.Cmd {
	groups, err := osauth.ListGroups(u.Username)
	if err != nil {
		groups = []uint32{}
	}

	nscommand, _ := nsenterCommandWrapper(u.UID, u.GID, groups, u.HomeDir, command...)

	cmd := exec.Command(nscommand[0], nscommand[1:]...) //nolint:noctx,gosec
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

	return cmd
}

func getWrappedCommand(nsArgs []string, uid, gid uint32, groups []uint32, home string) []string {
	gids := []string{}
	for _, g := range groups {
		gids = append(gids, strconv.Itoa(int(g)))
	}

	setPrivCmd := []string{
		"/bin/setpriv",
		"--groups=" + strings.Join(gids, ","),
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

	nsenterCmd = append(
		nsenterCmd,
		[]string{
			"-S",
			strconv.Itoa(int(uid)),
			"--wdns=" + home,
		}...,
	)

	return append(setPrivCmd, nsenterCmd...)
}

func nsenterCommandWrapper(uid, gid uint32, groups []uint32, home string, command ...string) ([]string, error) {
	if _, err := statFn("/usr/bin/nsenter"); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	namespaces := map[string]string{
		"mnt":    "-m",
		"uts":    "-u",
		"ipc":    "-i",
		"net":    "-n",
		"pid":    "-p",
		"cgroup": "-C",
	}

	present := map[string]string{}

	for ns, flag := range namespaces {
		if _, err := statFn("/proc/1/ns/" + ns); err != nil {
			continue
		}

		present[ns] = flag
	}

	args := nsenterArgs(present)

	return append(getWrappedCommand(args, uid, gid, groups, home), command...), nil
}

// SFTPServerCommand creates the command used by agent to start the SFTP server used in a SFTP connection.
func SFTPServerCommand() *exec.Cmd {
	return exec.Command("/proc/self/exe", []string{"sftp", string(SFTPServerModeDocker)}...) //nolint:noctx,gosec
}
