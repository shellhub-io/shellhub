package containers

import (
	"context"
	"io"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	dockerclient "github.com/docker/docker/client"
)

var _ Containers = new(DockerContainers)

// DockerContainers is a struct that represents a connector that uses Docker as the container runtime.
type DockerContainers struct {
	// cli is the Docker client.
	cli *dockerclient.Client
}

func NewDockerConnectorWithClient(cli *dockerclient.Client, server string, tenant string, privateKey string) (Containers, error) {
	return &DockerContainers{
		cli: cli,
	}, nil
}

// NewDockerConnector creates a new [Containers] that uses Docker as the container runtime.
func NewDockerConnector() (Containers, error) {
	cli, err := dockerclient.NewClientWithOpts(dockerclient.FromEnv, dockerclient.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}

	return &DockerContainers{
		cli: cli,
	}, nil
}

func (d *DockerContainers) List(ctx context.Context, opts ListOptions) ([]types.Container, error) {
	return d.cli.ContainerList(ctx, container.ListOptions{
		Size:   opts.Size,
		All:    opts.All,
		Latest: opts.Latest,
		Since:  opts.Since,
		Before: opts.Before,
		Limit:  opts.Limit,
	})
}

func (d *DockerContainers) Info(ctx context.Context, id string) (*types.ContainerJSON, error) {
	info, err := d.cli.ContainerInspect(ctx, id)
	if err != nil {
		return nil, err
	}

	return &info, nil
}

func (d *DockerContainers) Exec(ctx context.Context, id string, opts ExecOptions) (*types.HijackedResponse, error) {
	config := container.ExecOptions{
		Cmd:          strings.Split(opts.Cmd, " "),
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          opts.TTY,
	}

	execIDResp, err := d.cli.ContainerExecCreate(ctx, id, config)
	if err != nil {
		return nil, err
	}

	var consoleSize *[2]uint
	if opts.Height > 0 && opts.Width > 0 {
		consoleSize = &[2]uint{opts.Height, opts.Width}
	}

	resp, err := d.cli.ContainerExecAttach(ctx, execIDResp.ID, container.ExecAttachOptions{
		Tty:         opts.TTY,
		ConsoleSize: consoleSize,
	})
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func (d *DockerContainers) Start(ctx context.Context, id string, opts StartOptions) error {
	return d.cli.ContainerStart(ctx, id, container.StartOptions{})
}

func (d *DockerContainers) Stop(ctx context.Context, id string, opts StopOptions) error {
	return d.cli.ContainerStop(ctx, id, container.StopOptions{
		Timeout: opts.Timeout,
	})
}

func (d *DockerContainers) Restart(ctx context.Context, id string, opts RestartOptions) error {
	return d.cli.ContainerRestart(ctx, id, container.StopOptions{
		Timeout: opts.Timeout,
	})
}

func (d *DockerContainers) Kill(ctx context.Context, id string, opts KillOptions) error {
	return d.cli.ContainerKill(ctx, id, opts.Signal)
}

func (d *DockerContainers) Remove(ctx context.Context, id string, opts RemoveOptions) error {
	return d.cli.ContainerRemove(ctx, id, container.RemoveOptions{
		Force:         opts.Force,
		RemoveVolumes: opts.RemoveVolumes,
		RemoveLinks:   opts.RemoveLinks,
	})
}

func (d *DockerContainers) Pause(ctx context.Context, id string) error {
	return d.cli.ContainerPause(ctx, id)
}

func (d *DockerContainers) Unpause(ctx context.Context, id string) error {
	return d.cli.ContainerUnpause(ctx, id)
}

func (d *DockerContainers) Logs(ctx context.Context, id string, opts LogsOptions) (io.ReadCloser, error) {
	return d.cli.ContainerLogs(ctx, id, container.LogsOptions{
		Follow:     opts.Follow,
		ShowStdout: true, // NOTE: We always show stdout logs.
		ShowStderr: opts.Stderr,
		Since:      opts.Since,
		Until:      opts.Until,
		Timestamps: opts.Timestamps,
		Tail:       opts.Tail,
		Details:    opts.Details,
	})
}

func (d *DockerContainers) Stats(ctx context.Context, id string, opts StatsOptions) (container.StatsResponseReader, error) {
	resp, err := d.cli.ContainerStats(ctx, id, opts.Stream)
	if err != nil {
		return container.StatsResponseReader{}, err
	}

	return resp, nil
}

func (d *DockerContainers) Top(ctx context.Context, id string, opts TopOptions) (container.ContainerTopOKBody, error) {
	return d.cli.ContainerTop(ctx, id, []string{opts.PsArgs})
}

func (d *DockerContainers) Changes(ctx context.Context, id string) ([]container.FilesystemChange, error) {
	changes, err := d.cli.ContainerDiff(ctx, id)
	if err != nil {
		return nil, err
	}

	result := make([]container.FilesystemChange, len(changes))
	for i, change := range changes {
		result[i] = container.FilesystemChange{
			Path: change.Path,
			Kind: change.Kind,
		}
	}

	return result, nil
}

// Container file operations
func (d *DockerContainers) CopyToContainer(ctx context.Context, id string, path string, content io.Reader, opts CopyOptions) error {
	return d.cli.CopyToContainer(ctx, id, path, content, container.CopyToContainerOptions{
		AllowOverwriteDirWithFile: opts.AllowOverwrite,
		CopyUIDGID:                opts.CopyUIDGID,
	})
}

func (d *DockerContainers) CopyFromContainer(ctx context.Context, id string, path string) (io.ReadCloser, container.PathStat, error) {
	return d.cli.CopyFromContainer(ctx, id, path)
}

// Container management operations
func (d *DockerContainers) Rename(ctx context.Context, id string, opts RenameOptions) error {
	return d.cli.ContainerRename(ctx, id, opts.Name)
}

func (d *DockerContainers) Update(ctx context.Context, id string, opts UpdateOptions) (container.ContainerUpdateOKBody, error) {
	return d.cli.ContainerUpdate(ctx, id, container.UpdateConfig{
		Resources: container.Resources{
			CPUShares:          opts.CPUShares,
			Memory:             opts.Memory,
			CgroupParent:       opts.CgroupParent,
			BlkioWeight:        opts.BlkioWeight,
			CPUPeriod:          opts.CPUPeriod,
			CPUQuota:           opts.CPUQuota,
			CPURealtimePeriod:  opts.CPURealtimePeriod,
			CPURealtimeRuntime: opts.CPURealtimeRuntime,
			CpusetCpus:         opts.CpusetCpus,
			CpusetMems:         opts.CpusetMems,
			KernelMemory:       opts.KernelMemory,
			MemoryReservation:  opts.MemoryReservation,
			MemorySwap:         opts.MemorySwap,
			MemorySwappiness:   opts.MemorySwappiness,
			OomKillDisable:     opts.OomKillDisable,
			PidsLimit:          opts.PidsLimit,
		},
		RestartPolicy: opts.RestartPolicy,
	})
}

func (d *DockerContainers) Resize(ctx context.Context, id string, opts ResizeOptions) error {
	return d.cli.ContainerResize(ctx, id, container.ResizeOptions{
		Height: opts.Height,
		Width:  opts.Width,
	})
}

func (d *DockerContainers) Commit(ctx context.Context, id string, opts CommitOptions) (types.IDResponse, error) {
	return d.cli.ContainerCommit(ctx, id, container.CommitOptions{
		Reference: opts.Repo + ":" + opts.Tag,
		Comment:   opts.Comment,
		Author:    opts.Author,
		Pause:     opts.Pause,
		Changes:   opts.Changes,
		Config:    opts.Config,
	})
}

func (d *DockerContainers) Wait(ctx context.Context, id string, opts WaitOptions) (<-chan container.WaitResponse, <-chan error) {
	condition := container.WaitConditionNotRunning
	if opts.Condition != "" {
		condition = container.WaitCondition(opts.Condition)
	}

	return d.cli.ContainerWait(ctx, id, condition)
}

func (d *DockerContainers) Export(ctx context.Context, id string) (io.ReadCloser, error) {
	return d.cli.ContainerExport(ctx, id)
}
