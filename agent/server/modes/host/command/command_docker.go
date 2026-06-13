//go:build docker
// +build docker

package command

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/shellhub-io/shellhub/agent/pkg/osauth"
	log "github.com/sirupsen/logrus"
)

// statFn is a seam for os.Stat used when probing /proc/1/ns/* entries.
// It can be replaced in tests to avoid filesystem access.
var statFn = os.Stat

// probeTimeNSJoinable checks whether the agent can join the host's time
// namespace by running a dry-run nsenter command with a short timeout.
//
// Return values:
//   - (true,  true):  the host time namespace is joinable (exit 0); cache it.
//   - (false, true):  definitive denial (EPERM/EACCES, i.e. *exec.ExitError
//     with code other than 126/127); result is safe to cache.
//   - (false, false): nsenter or /bin/true not found (exit 126/127 or
//     os.ErrNotExist); logged as a distinct warning; must NOT be cached.
//   - (false, false): context deadline exceeded or any other transient error;
//     must NOT be cached.
func probeTimeNSJoinable() (result bool, definitive bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	err := exec.CommandContext(ctx, "/usr/bin/nsenter", "-t", "1", "-T", "--", "/bin/true").Run() //nolint:gosec
	if err == nil {
		return true, true
	}

	// os.ErrNotExist means the nsenter binary itself was not found at the
	// absolute path (/usr/bin/nsenter). When an absolute path is given,
	// exec.Command skips LookPath entirely, so exec.ErrNotFound is never
	// returned; a missing binary surfaces as an *fs.PathError wrapping ENOENT.
	if errors.Is(err, os.ErrNotExist) {
		log.WithError(err).Warn("nsenter not found; time namespace join probe skipped")

		return false, false
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		code := exitErr.ExitCode()

		if code == 126 || code == 127 {
			// 126 = command cannot execute; 127 = command not found.
			// These are environment issues, not a kernel denial — do not cache.
			log.WithError(err).WithField("exit_code", code).
				Warn("time namespace probe returned 126/127; skipping -T flag")

			return false, false
		}

		// Any other non-zero exit (e.g. EPERM → 1) is a definitive denial.
		return false, true
	}

	// Context deadline or other transient error — do not cache.
	return false, false
}

// newTimeNSMemoizer wraps a probe function in a tri-state memoizer.
//
// State is held as a *bool:
//   - nil  → undetermined; the probe has not yet produced a definitive result.
//   - non-nil → a definitive result has been cached; the stored value is returned
//     on every subsequent call without invoking the probe again.
//
// The probe is called outside the lock so slow probes never serialize concurrent
// callers.  Only definitive results (where probe returns definitive=true) are
// stored back under the lock.
//
// A separate sync.Once ensures the "time namespace not joinable" warning is
// emitted at most once per process lifetime.
func newTimeNSMemoizer(probe func() (bool, bool)) func() bool {
	var (
		mu       sync.Mutex
		cached   *bool
		warnOnce sync.Once
	)

	return func() bool {
		// Fast path: read cached result under lock.
		mu.Lock()
		if cached != nil {
			v := *cached
			mu.Unlock()

			return v
		}

		mu.Unlock()

		// Slow path: run the probe outside the lock.
		result, definitive := probe()

		if definitive {
			mu.Lock()
			// Only store if still undetermined (another goroutine may have
			// raced and stored first — both would get the same definitive
			// answer, so either value is fine; we keep whichever got there
			// first).
			if cached == nil {
				v := result
				cached = &v
			}

			mu.Unlock()

			if !result {
				warnOnce.Do(func() {
					log.Warn("time namespace not joinable; continuing without -T")
				})
			}
		}

		return result
	}
}

// timeNSJoinableFn is the seam used by nsenterCommandWrapper to decide whether
// to pass -T to nsenter.  The default value wraps the real probe in the
// tri-state memoizer so that only definitive results are cached and transient
// or misconfiguration failures are retried on the next call.
//
// Tests replace this variable with a plain func() bool stub.
var timeNSJoinableFn = newTimeNSMemoizer(probeTimeNSJoinable)

// nsenterArgs builds the nsenter flag slice from the present namespace map.
// Non-time namespace flags are taken directly from present. The -T flag is
// appended only when joinTime is true.
func nsenterArgs(present map[string]string, joinTime bool) []string {
	args := []string{}

	for _, flag := range present {
		args = append(args, flag)
	}

	if joinTime {
		args = append(args, "-T")
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
// It probes /proc/1/ns/* for all namespaces except time using statFn, then
// delegates flag assembly to nsenterArgs, passing the result of timeNSJoinableFn
// to decide whether to include the -T (time namespace) flag.
func nsenterCommandWrapper(uid, gid uint32, groups []uint32, home string, command ...string) ([]string, error) {
	if _, err := statFn("/usr/bin/nsenter"); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	// Namespaces to probe — time is excluded here and handled via joinTime in nsenterArgs.
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

	args := nsenterArgs(present, timeNSJoinableFn())

	return append(getWrappedCommand(args, uid, gid, groups, home), command...), nil
}

// SFTPServerCommand creates the command used by agent to start the SFTP server used in a SFTP connection.
func SFTPServerCommand() *exec.Cmd {
	return exec.Command("/proc/self/exe", []string{"sftp", string(SFTPServerModeDocker)}...) //nolint:gosec
}
