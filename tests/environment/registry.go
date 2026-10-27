package environment

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/moby/moby/client"
	"github.com/shellhub-io/shellhub/pkg/testimage"
	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	registryImage = "registry:2"
	registryPort  = "5000/tcp"

	anonymousRegistryAuth = "e30="
)

// Registry is a Docker registry the run's daemon reaches at Address, localhost and the port the
// registry is published on. The daemon talks plain HTTP to a registry on localhost, so it pushes
// and pulls there without TLS or credentials.
type Registry struct {
	Address string

	run       *Run
	container tc.Container
}

// StartRegistry starts a registry labelled with run and returns it once it answers. It returns an
// error when the registry image cannot be pulled, the container cannot start or does not answer,
// or its published port cannot be read; the container is removed on every error after it started.
// Close removes it.
func StartRegistry(ctx context.Context, run *Run) (*Registry, error) {
	container, err := tc.GenericContainer(ctx, tc.GenericContainerRequest{
		ContainerRequest: tc.ContainerRequest{
			Image:        testimage.DockerHub(registryImage),
			ExposedPorts: []string{registryPort},
			Labels:       run.Labels(),
			WaitingFor:   wait.ForHTTP("/v2/").WithPort(registryPort),
		},
		Started: true,
		Logger:  log.New(io.Discard, "", log.LstdFlags),
	})
	if err != nil {
		if container != nil {
			_ = container.Terminate(context.WithoutCancel(ctx))
		}

		return nil, fmt.Errorf("start the registry: %w", err)
	}

	port, err := container.MappedPort(ctx, registryPort)
	if err != nil {
		_ = container.Terminate(context.WithoutCancel(ctx))

		return nil, fmt.Errorf("read the registry's port: %w", err)
	}

	return &Registry{Address: "localhost:" + port.Port(), run: run, container: container}, nil
}

// Push tags the local image as repository:tag in the registry, pushes it, and removes that tag
// from the daemon again, so the image under it can only be had by pulling from the registry. It
// returns the reference the image was pushed as, or an error when tagging, pushing or removing
// the tag fails, carrying the daemon's message.
func (r *Registry) Push(ctx context.Context, image, repository, tag string) (string, error) {
	reference := r.Address + "/" + repository + ":" + tag

	if _, err := r.run.client.ImageTag(ctx, client.ImageTagOptions{Source: image, Target: reference}); err != nil {
		return "", fmt.Errorf("tag %s as %s: %w", image, reference, err)
	}

	pushed, err := r.run.client.ImagePush(ctx, reference, client.ImagePushOptions{RegistryAuth: anonymousRegistryAuth})
	if err != nil {
		return "", fmt.Errorf("push %s: %w", reference, err)
	}

	if err := pushed.Wait(ctx); err != nil {
		return "", fmt.Errorf("push %s: %w", reference, err)
	}

	if _, err := r.run.client.ImageRemove(ctx, reference, client.ImageRemoveOptions{}); err != nil {
		return "", fmt.Errorf("untag %s: %w", reference, err)
	}

	return reference, nil
}

// Close removes the registry and everything pushed to it, on a context that outlives ctx's
// cancellation, so it can sit in a cleanup. It returns an error when the registry's container
// cannot be stopped or removed.
func (r *Registry) Close(ctx context.Context) error {
	return r.container.Terminate(context.WithoutCancel(ctx))
}
