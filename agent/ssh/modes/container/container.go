package container

import (
	"context"

	"github.com/docker/docker/api/types"
	dockercontainer "github.com/docker/docker/api/types/container"
	dockerclient "github.com/docker/docker/client"
	"github.com/docker/docker/pkg/process"
	"github.com/shellhub-io/shellhub/agent/auth"
	"github.com/shellhub-io/shellhub/agent/ssh/modes"
)

type Mode struct {
	modes.Authenticator
	modes.Sessioner
}

func attachShellToContainer(ctx context.Context, cli dockerclient.APIClient, container string, user *auth.User, size [2]int) (*types.HijackedResponse, string, error) {
	return attachToContainer(ctx, cli, "shell", container, user, true, []string{}, size)
}

func attachExecToContainer(ctx context.Context, cli dockerclient.APIClient, container string, user *auth.User, isPty bool, commands []string, size [2]int) (*types.HijackedResponse, string, error) {
	return attachToContainer(ctx, cli, "exec", container, user, isPty, commands, size)
}

func attachHereDocToContainer(ctx context.Context, cli dockerclient.APIClient, container string, user *auth.User, size [2]int) (*types.HijackedResponse, string, error) {
	return attachToContainer(ctx, cli, "heredoc", container, user, false, []string{}, size)
}

func attachToContainer(ctx context.Context, cli dockerclient.APIClient, requestType string, container string, user *auth.User, isPty bool, commands []string, size [2]int) (*types.HijackedResponse, string, error) {
	if user.Shell == "" {
		user.Shell = "/bin/sh"
	}

	s := &[2]uint{
		uint(size[0]), //nolint:gosec
		uint(size[1]), //nolint:gosec
	}

	id, err := cli.ContainerExecCreate(ctx, container, dockercontainer.ExecOptions{
		User:         user.Username,
		Tty:          isPty,
		ConsoleSize:  s,
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Cmd: func() []string {
			switch requestType {
			case "shell":
				return []string{user.Shell}
			case "exec":
				// NOTE: When the exec session's has `-t` or `-tt` flag, the command must be executed into a tty/pty.
				// the Shell's `-c` flag is used to do this.
				if isPty {
					return append([]string{user.Shell, "-c"}, commands...)
				}

				return commands
			case "heredoc":
				return []string{user.Shell}
			default:
				return []string{}
			}
		}(),
	})
	if err != nil {
		return nil, "", err
	}

	res, err := cli.ContainerExecAttach(ctx, id.ID, dockercontainer.ExecStartOptions{
		Tty:         isPty,
		ConsoleSize: s,
	})

	return &res, id.ID, err
}

func exitCodeExecFromContainer(cli dockerclient.APIClient, id string) (int, error) {
	inspected, err := cli.ContainerExecInspect(context.Background(), id)
	if err != nil {
		return -1, err
	}

	if inspected.Running {
		// NOTE: When a process is running after the exec command, it is necessary to kill it.
		return 0, process.Kill(inspected.Pid)
	}

	return inspected.ExitCode, nil
}
