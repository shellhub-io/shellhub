//go:build !freebsd

package host

import (
	"fmt"
	"os"
	"os/exec"

	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/agent/pkg/osauth"
	"github.com/shellhub-io/shellhub/agent/server/modes/host/command"
)

func generateShellCmd(deviceName string, session gliderssh.Session, term string) *exec.Cmd {
	username := session.User()
	envs := session.Environ()

	user, err := osauth.LookupUser(username)
	if err != nil {
		return nil
	}

	shell := user.Shell
	if shell == "" {
		shell = os.Getenv("SHELL")
	}

	if term == "" {
		term = "xterm"
	}

	authSock := session.Context().Value("SSH_AUTH_SOCK")
	if authSock != nil {
		envs = append(envs, fmt.Sprintf("%s=%s", "SSH_AUTH_SOCK", authSock.(string)))
	}

	cmd := command.NewCmd(user, shell, term, deviceName, envs, shell, "--login")

	return cmd
}
