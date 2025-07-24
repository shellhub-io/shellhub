package docker

import (
	"context"
	"io"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	dockerclient "github.com/docker/docker/client"
	"github.com/shellhub-io/shellhub/agent/ssh/modes/container/engines"
)

var _ engines.Engine = new(DockerEngine)

// DockerEngine is a struct that represents a connector that uses Docker as the container runtime.
type DockerEngine struct {
	// cli is the Docker client.
	cli *dockerclient.Client
}

func NewDockerConnectorWithClient(cli *dockerclient.Client, server string, tenant string, privateKey string) (engines.Engine, error) {
	return &DockerEngine{
		cli: cli,
	}, nil
}

func NewDockerEngine() (engines.Engine, error) {
	cli, err := dockerclient.NewClientWithOpts(dockerclient.FromEnv, dockerclient.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}

	return &DockerEngine{
		cli: cli,
	}, nil
}

func (d *DockerEngine) List(ctx context.Context, opts engines.ListOptions) ([]types.Container, error) {
	return d.cli.ContainerList(ctx, container.ListOptions{
		Size:   opts.Size,
		All:    opts.All,
		Latest: opts.Latest,
		Since:  opts.Since,
		Before: opts.Before,
		Limit:  opts.Limit,
	})
}

func (d *DockerEngine) Info(ctx context.Context, id string) (*types.ContainerJSON, error) {
	info, err := d.cli.ContainerInspect(ctx, id)
	if err != nil {
		return nil, err
	}

	return &info, nil
}

func (d *DockerEngine) Exec(ctx context.Context, id string, opts engines.ExecOptions) (*types.HijackedResponse, error) {
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

func (d *DockerEngine) Start(ctx context.Context, id string, opts engines.StartOptions) error {
	return d.cli.ContainerStart(ctx, id, container.StartOptions{})
}

func (d *DockerEngine) Stop(ctx context.Context, id string, opts engines.StopOptions) error {
	return d.cli.ContainerStop(ctx, id, container.StopOptions{
		Timeout: opts.Timeout,
	})
}

func (d *DockerEngine) Restart(ctx context.Context, id string, opts engines.RestartOptions) error {
	return d.cli.ContainerRestart(ctx, id, container.StopOptions{
		Timeout: opts.Timeout,
	})
}

func (d *DockerEngine) Kill(ctx context.Context, id string, opts engines.KillOptions) error {
	return d.cli.ContainerKill(ctx, id, opts.Signal)
}

func (d *DockerEngine) Remove(ctx context.Context, id string, opts engines.RemoveOptions) error {
	return d.cli.ContainerRemove(ctx, id, container.RemoveOptions{
		Force:         opts.Force,
		RemoveVolumes: opts.RemoveVolumes,
		RemoveLinks:   opts.RemoveLinks,
	})
}

func (d *DockerEngine) Pause(ctx context.Context, id string) error {
	return d.cli.ContainerPause(ctx, id)
}

func (d *DockerEngine) Unpause(ctx context.Context, id string) error {
	return d.cli.ContainerUnpause(ctx, id)
}

func (d *DockerEngine) Logs(ctx context.Context, id string, opts engines.LogsOptions) (io.ReadCloser, error) {
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

func (d *DockerEngine) Stats(ctx context.Context, id string, opts engines.StatsOptions) (container.StatsResponseReader, error) {
	resp, err := d.cli.ContainerStats(ctx, id, opts.Stream)
	if err != nil {
		return container.StatsResponseReader{}, err
	}

	return resp, nil
}

func (d *DockerEngine) Top(ctx context.Context, id string, opts engines.TopOptions) (container.ContainerTopOKBody, error) {
	return d.cli.ContainerTop(ctx, id, []string{opts.PsArgs})
}

func (d *DockerEngine) Changes(ctx context.Context, id string) ([]container.FilesystemChange, error) {
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

func (d *DockerEngine) CopyToContainer(ctx context.Context, id string, path string, content io.Reader, opts engines.CopyOptions) error {
	return d.cli.CopyToContainer(ctx, id, path, content, container.CopyToContainerOptions{
		AllowOverwriteDirWithFile: opts.AllowOverwrite,
		CopyUIDGID:                opts.CopyUIDGID,
	})
}

func (d *DockerEngine) CopyFromContainer(ctx context.Context, id string, path string) (io.ReadCloser, container.PathStat, error) {
	return d.cli.CopyFromContainer(ctx, id, path)
}

func (d *DockerEngine) Rename(ctx context.Context, id string, opts engines.RenameOptions) error {
	return d.cli.ContainerRename(ctx, id, opts.Name)
}

func (d *DockerEngine) Update(ctx context.Context, id string, opts engines.UpdateOptions) (container.ContainerUpdateOKBody, error) {
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

func (d *DockerEngine) Resize(ctx context.Context, id string, opts engines.ResizeOptions) error {
	return d.cli.ContainerResize(ctx, id, container.ResizeOptions{
		Height: opts.Height,
		Width:  opts.Width,
	})
}

func (d *DockerEngine) Commit(ctx context.Context, id string, opts engines.CommitOptions) (types.IDResponse, error) {
	return d.cli.ContainerCommit(ctx, id, container.CommitOptions{
		Reference: opts.Repo + ":" + opts.Tag,
		Comment:   opts.Comment,
		Author:    opts.Author,
		Pause:     opts.Pause,
		Changes:   opts.Changes,
		Config:    opts.Config,
	})
}

func (d *DockerEngine) Wait(ctx context.Context, id string, opts engines.WaitOptions) (<-chan container.WaitResponse, <-chan error) {
	condition := container.WaitConditionNotRunning
	if opts.Condition != "" {
		condition = container.WaitCondition(opts.Condition)
	}

	return d.cli.ContainerWait(ctx, id, condition)
}

func (d *DockerEngine) Export(ctx context.Context, id string) (io.ReadCloser, error) {
	return d.cli.ContainerExport(ctx, id)
}
