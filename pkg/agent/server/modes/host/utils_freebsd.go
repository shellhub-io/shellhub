//go:build freebsd

package host

import (
	"os"
	"os/exec"

	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/pkg/agent/pkg/osauth"
	"github.com/shellhub-io/shellhub/pkg/agent/server/modes/host/command"
)

func generateShellCmd(deviceName string, session gliderssh.Session, term string) *exec.Cmd {
	username := session.User()
	envs := session.Environ()

	shell := os.Getenv("SHELL")

	user, err := osauth.LookupUser(username)
	if err != nil {
		return nil
	}

	if shell == "" {
		shell = user.Shell
	}

	if term == "" {
		term = "xterm"
	}

	cmd := command.NewCmd(user, shell, term, deviceName, envs, shell, "-")

	return cmd
}
