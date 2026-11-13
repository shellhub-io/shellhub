package connector

import (
	"context"
	"errors"
	"testing"

	dockercontainer "github.com/docker/docker/api/types/container"
	dockerclient "github.com/docker/docker/client"
	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/agent/pkg/osauth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errExecRefused = errors.New("exec refused by the fake docker client")

type recordingDockerClient struct {
	dockerclient.APIClient
	cmd []string
}

func (c *recordingDockerClient) ContainerExecCreate(_ context.Context, _ string, options dockercontainer.ExecOptions) (dockercontainer.ExecCreateResponse, error) {
	c.cmd = options.Cmd

	return dockercontainer.ExecCreateResponse{}, errExecRefused
}

type userContext struct {
	gliderssh.Context
	user *osauth.User
}

func (c *userContext) Value(key any) any {
	if key == "user" {
		return c.user
	}

	return nil
}

type execSession struct {
	gliderssh.Session
	ctx        *userContext
	isPty      bool
	rawCommand string
	command    []string
}

func (s *execSession) Context() gliderssh.Context { return s.ctx }

func (s *execSession) RawCommand() string { return s.rawCommand }

func (s *execSession) Command() []string { return s.command }

func (s *execSession) Pty() (gliderssh.Pty, <-chan gliderssh.Window, bool) {
	return gliderssh.Pty{}, nil, s.isPty
}

func TestExec_RunsTheCommandThroughTheLoginShell(t *testing.T) {
	cases := []struct {
		name  string
		isPty bool
	}{
		{name: "without a pty", isPty: false},
		{name: "with a pty", isPty: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			docker := &recordingDockerClient{}
			container := "container"
			s := NewSessioner(&container, docker)

			err := s.Exec(&execSession{
				ctx:        &userContext{user: &osauth.User{Username: "git", Shell: "/usr/bin/git-shell"}},
				isPty:      tc.isPty,
				rawCommand: "touch /tmp/hook && echo $HOME",
				command:    []string{"touch", "/tmp/hook", "&&", "echo", "$HOME"},
			})
			require.ErrorIs(t, err, errExecRefused)

			assert.Equal(t, []string{"/usr/bin/git-shell", "-c", "touch /tmp/hook && echo $HOME"}, docker.cmd)
		})
	}
}

func TestExec_FallsBackToShWhenTheAccountHasNoShell(t *testing.T) {
	docker := &recordingDockerClient{}
	container := "container"
	s := NewSessioner(&container, docker)

	err := s.Exec(&execSession{
		ctx:        &userContext{user: &osauth.User{Username: "app"}},
		rawCommand: "id",
		command:    []string{"id"},
	})
	require.ErrorIs(t, err, errExecRefused)

	assert.Equal(t, []string{"/bin/sh", "-c", "id"}, docker.cmd)
}
