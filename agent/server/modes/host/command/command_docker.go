//go:build docker
// +build docker

package command

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/shellhub-io/shellhub/agent/pkg/osauth"
)

// statFn is a seam for os.Stat used when probing /proc/1/ns/* entries.
// It can be replaced in tests to avoid filesystem access.
var statFn = os.Stat

// nsenterArgs builds the nsenter flag slice from the present namespace map.
// Docker always shares the host time namespace, so -T is never passed.
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

	// NOTE: Wrap the command with nsenter and setpriv to run it inside the
	// host's namespaces with the correct user and groups. This is necessary
	// because the agent is running inside a Docker container and we want to
	// execute the command in the host's context.
	nscommand, _ := nsenterCommandWrapper(u.UID, u.GID, groups, u.HomeDir, command...)

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

func getWrappedCommand(nsArgs []string, uid, gid uint32, groups []uint32, home string) []string {
	gids := []string{}
	for _, g := range groups {
		gids = append(gids, strconv.Itoa(int(g)))
	}

	setPrivCmd := []string{
		"/bin/setpriv",
		fmt.Sprintf("--groups=%s", strings.Join(gids, ",")),
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
			fmt.Sprintf("--wdns=%s", home),
		}...,
	)

	return append(setPrivCmd, nsenterCmd...)
}

// nsenterCommandWrapper builds the full nsenter+setpriv command slice.
// It probes /proc/1/ns/* for each namespace using statFn, then delegates
// flag assembly to nsenterArgs. The time namespace is never joined because
// Docker always shares the host time namespace.
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
		if _, err := statFn(fmt.Sprintf("/proc/1/ns/%s", ns)); err != nil {
			continue
		}

		present[ns] = flag
	}

	args := nsenterArgs(present)

	return append(getWrappedCommand(args, uid, gid, groups, home), command...), nil
}

// SFTPServerCommand creates the command used by agent to start the SFTP server used in a SFTP connection.
func SFTPServerCommand() *exec.Cmd {
	return exec.Command("/proc/self/exe", []string{"sftp", string(SFTPServerModeDocker)}...) //nolint:gosec
}
