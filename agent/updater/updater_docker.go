//go:build docker
// +build docker

package selfupdater

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/docker/docker/api/types"
	containertypes "github.com/docker/docker/api/types/container"
	dockerimage "github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/shellhub-io/shellhub/pkg/dockerutils"
)

type dockerContainer struct {
	info *types.ContainerJSON
}

func (c *dockerContainer) splitImageVersion() (image, version string) {
	parts := strings.SplitN(c.info.Config.Image, ":", 2)
	image, version = parts[0], ""

	if len(parts) == 2 {
		version = parts[1]
	}

	return
}

type dockerUpdater struct {
	api client.CommonAPIClient
}

func (d *dockerUpdater) CurrentVersion() (*semver.Version, error) {
	container, err := d.currentContainer()
	if err != nil {
		return nil, err
	}

	_, version := container.splitImageVersion()

	return semver.NewVersion(version)
}

func (d *dockerUpdater) ApplyUpdate(v *semver.Version) error {
	container, err := d.currentContainer()
	if err != nil {
		return err
	}

	image, _ := container.splitImageVersion()
	_, err = d.updateContainer(container, fmt.Sprintf("%s:%s", image, v.Original()), "", true)

	return err
}

func (d *dockerUpdater) CompleteUpdate() error {
	container, err := d.currentContainer()
	if err != nil {
		return err
	}

	parent, err := d.parentContainer()
	if err != nil {
		return err
	}

	if parent == nil {
		return nil
	}

	if err := d.stopContainer(parent); err != nil {
		return err
	}

	_, pv := parent.splitImageVersion()
	v, _ := semver.NewVersion(pv)
	v0_4_0, _ := semver.NewVersion("v0.4.0")

	// Append /dev to mount if old container version is less than v0.4.0
	// since /dev from host is required inside container to mount a pseudo-tty
	if v.LessThan(v0_4_0) {
		parent.info.HostConfig.Mounts = []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: "/dev",
				Target: "/dev",
			},
		}
	}

	// Append /var/run and /var/log to mount if old container
	// version is less than v0.5.0 since utmp and wtmp are
	// required inside container to record login sessions
	v0_5_0, _ := semver.NewVersion("v0.5.0")
	if v.LessThan(v0_5_0) {
		parent.info.HostConfig.Mounts = []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: "/var/run",
				Target: "/var/run",
			},
			{
				Type:   mount.TypeBind,
				Source: "/var/log",
				Target: "/var/log",
			},
		}
	}

	// Append /etc/resolv.conf to mount if old container version
	// is less than v0.7.3 since /etc/resolv.conf is required
	// inside the container to update host networking after boot.
	v0_7_3, _ := semver.NewVersion("v0.7.3")
	if v.LessThan(v0_7_3) {
		parent.info.HostConfig.Mounts = []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: "/etc/resolv.conf",
				Target: "/etc/resolv.conf",
			},
		}
	}

	_, err = d.updateContainer(parent, container.info.Config.Image, parent.info.Name, false)
	if err != nil {
		return err
	}

	if err := d.stopContainer(container); err != nil {
		return err
	}

	return nil
}

func (d *dockerUpdater) getContainer(id string) (*dockerContainer, error) {
	ctx := context.Background()

	info, err := d.api.ContainerInspect(ctx, id)
	if err != nil {
		return nil, err
	}

	return &dockerContainer{info: &info}, nil
}

func (d *dockerUpdater) currentContainer() (*dockerContainer, error) {
	id, err := dockerutils.CurrentContainerID()
	if err != nil {
		return nil, err
	}

	return d.getContainer(id)
}

func (d *dockerUpdater) parentContainer() (*dockerContainer, error) {
	id := os.Getenv("PARENT_CONTAINER")
	if id == "" {
		return nil, nil
	}

	return d.getContainer(id)
}

func (d *dockerUpdater) stopContainer(container *dockerContainer) error {
	ctx := context.Background()

	timeout := 60 // seconds
	if err := d.api.ContainerStop(ctx, container.info.ID, containertypes.StopOptions{Timeout: &timeout}); err != nil {
		return err
	}

	opts := containertypes.RemoveOptions{Force: true, RemoveVolumes: true}
	err := d.api.ContainerRemove(ctx, container.info.ID, opts)

	return err
}

func (d *dockerUpdater) updateContainer(container *dockerContainer, image, name string, parent bool) (*dockerContainer, error) { //nolint:unparam
	ctx := context.Background()

	// Clone container container config and update the image name
	config := container.info.Config
	config.Image = image

	// Set current container as parent of the new container
	if parent {
		config.Env = replaceOrAppendEnvValues(config.Env, []string{fmt.Sprintf("PARENT_CONTAINER=%s", container.info.ID)})
	}

	netConfig := &network.NetworkingConfig{EndpointsConfig: container.info.NetworkSettings.Networks}

	rd, err := d.api.ImagePull(ctx, image, dockerimage.PullOptions{})
	if err != nil {
		return nil, err
	}
	defer rd.Close()

	// Wait for image to pull
	_, err = io.Copy(io.Discard, rd)
	if err != nil {
		return nil, err
	}

	// Create a new container using the cloned container config
	clone, err := d.api.ContainerCreate(ctx, config, container.info.HostConfig, netConfig, nil, name)
	if err != nil {
		return nil, err
	}

	if err := d.api.ContainerStart(ctx, clone.ID, containertypes.StartOptions{}); err != nil {
		return nil, err
	}

	return d.getContainer(clone.ID)
}

func NewUpdater(version string) (Updater, error) {
	// ensure we are running inside a docker container, otherwise returns a dummy updater implementation
	if _, err := os.Stat("/.dockerenv"); os.IsNotExist(err) {
		return &nativeUpdater{version}, nil
	}

	api, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, err
	}

	api.NegotiateAPIVersion(context.Background())

	return &dockerUpdater{api: api}, nil
}

func replaceOrAppendEnvValues(defaults, overrides []string) []string {
	cache := make(map[string]int, len(defaults))
	for i, e := range defaults {
		index := strings.Index(e, "=")
		cache[e[:index]] = i
	}

	for _, value := range overrides {
		// Values w/o = means they want this env to be removed/unset.
		index := strings.Index(value, "=")
		if index < 0 {
			// no "=" in value
			if i, exists := cache[value]; exists {
				defaults[i] = "" // Used to indicate it should be removed
			}

			continue
		}

		if i, exists := cache[value[:index]]; exists {
			defaults[i] = value
		} else {
			defaults = append(defaults, value)
		}
	}

	// Now remove all entries that we want to "unset"
	for i := 0; i < len(defaults); i++ {
		if defaults[i] == "" {
			defaults = append(defaults[:i], defaults[i+1:]...)
			i--
		}
	}

	return defaults
}
