package environment

import (
	"fmt"
	"os"
	"sync"

	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/flags"
)

var pinDockerHost = sync.OnceValue(func() error {
	if os.Getenv("DOCKER_HOST") != "" {
		return nil
	}

	cli, err := newDockerCli()
	if err != nil {
		return err
	}

	return os.Setenv("DOCKER_HOST", cli.DockerEndpoint().Host)
})

func newDockerCli() (*command.DockerCli, error) {
	cli, err := command.NewDockerCli()
	if err != nil {
		return nil, fmt.Errorf("new docker cli: %w", err)
	}

	if err := cli.Initialize(&flags.ClientOptions{}); err != nil {
		return nil, fmt.Errorf("initialize docker cli: %w", err)
	}

	return cli, nil
}
