package containers

import (
	"context"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
)

// Container is a struct that represents a container that will be managed by the connector.
type Container struct {
	// ID is the container ID.
	ID string
	// Name is the container name.
	Name string
	// ServerAddress is the ShellHub address of the server that the agent will connect to.
	ServerAddress string
	// Tenant is the tenant ID of the namespace that the agent belongs to.
	Tenant string
	// PrivateKey is the private key of the device. Specify the path to store the container private key. If not
	// provided, the agent will generate a new one. This is required.
	PrivateKey string
	// Cancel is a function that is used to stop the goroutine that is running the agent for this container.
	Cancel context.CancelFunc
}

// ExecOptions defines options for attaching to a container.
type ExecOptions struct {
	TTY    bool   `query:"tty"`
	Cmd    string `query:"cmd"`
	Width  uint   `query:"w"`
	Height uint   `query:"h"`
}

// ListOptions holds parameters to list containers with.
type ListOptions struct {
	Size   bool   `query:"size"`
	All    bool   `query:"all"`
	Latest bool   `query:"latest"`
	Since  string `query:"since"`
	Before string `query:"before"`
	Limit  int    `query:"limit"`
}

// StartOptions holds parameters to start a container.
type StartOptions struct {
	DetachKeys string `query:"detachKeys"`
}

// StopOptions holds parameters to stop a container.
type StopOptions struct {
	Timeout *int `query:"t"`
}

// RestartOptions holds parameters to restart a container.
type RestartOptions struct {
	Timeout *int `query:"t"`
}

// KillOptions holds parameters to kill a container.
type KillOptions struct {
	Signal string `query:"signal"`
}

// RemoveOptions holds parameters to remove a container.
type RemoveOptions struct {
	Force         bool `query:"force"`
	RemoveVolumes bool `query:"v"`
	RemoveLinks   bool `query:"link"`
}

// LogsOptions holds parameters to get container logs.
type LogsOptions struct {
	Follow bool `query:"follow"`
	// Stdout     bool   `query:"stdout"`
	Stderr     bool   `query:"stderr"`
	Since      string `query:"since"`
	Until      string `query:"until"`
	Timestamps bool   `query:"timestamps"`
	Tail       string `query:"tail"`
	Details    bool   `query:"details"`
}

// StatsOptions holds parameters to get container stats.
type StatsOptions struct {
	Stream   bool `query:"stream"`
	OneShot  bool `query:"one-shot"`
	NoStream bool `query:"no-stream"`
}

// TopOptions holds parameters to get container processes.
type TopOptions struct {
	PsArgs string `query:"ps_args"`
}

// AttachOptions holds parameters to attach to a container.
type AttachOptions struct {
	Stdin  bool `query:"stdin"`
	Stdout bool `query:"stdout"`
	Stderr bool `query:"stderr"`
	Stream bool `query:"stream"`
	Logs   bool `query:"logs"`
}

// CopyOptions holds parameters for copy operations.
type CopyOptions struct {
	Path           string `query:"path"`
	NoOverwrite    bool   `query:"noOverwriteDirNonDir"`
	CopyUIDGID     bool   `query:"copyUIDGID"`
	AllowOverwrite bool   `query:"allowOverwriteDirWithFile"`
}

// RenameOptions holds parameters to rename a container.
type RenameOptions struct {
	Name string `query:"name"`
}

// UpdateOptions holds parameters to update a container.
type UpdateOptions struct {
	CPUShares          int64                   `json:"CpuShares"`
	Memory             int64                   `json:"Memory"`
	CgroupParent       string                  `json:"CgroupParent"`
	BlkioWeight        uint16                  `json:"BlkioWeight"`
	CPUPeriod          int64                   `json:"CpuPeriod"`
	CPUQuota           int64                   `json:"CpuQuota"`
	CPURealtimePeriod  int64                   `json:"CpuRealtimePeriod"`
	CPURealtimeRuntime int64                   `json:"CpuRealtimeRuntime"`
	CpusetCpus         string                  `json:"CpusetCpus"`
	CpusetMems         string                  `json:"CpusetMems"`
	KernelMemory       int64                   `json:"KernelMemory"`
	MemoryReservation  int64                   `json:"MemoryReservation"`
	MemorySwap         int64                   `json:"MemorySwap"`
	MemorySwappiness   *int64                  `json:"MemorySwappiness"`
	OomKillDisable     *bool                   `json:"OomKillDisable"`
	PidsLimit          *int64                  `json:"PidsLimit"`
	RestartPolicy      container.RestartPolicy `json:"RestartPolicy"`
}

// ResizeOptions holds parameters to resize a container TTY.
type ResizeOptions struct {
	Height uint `query:"h"`
	Width  uint `query:"w"`
}

// CommitOptions holds parameters to commit a container.
type CommitOptions struct {
	Container string            `json:"container"`
	Repo      string            `json:"repo"`
	Tag       string            `json:"tag"`
	Comment   string            `json:"comment"`
	Author    string            `json:"author"`
	Pause     bool              `json:"pause"`
	Changes   []string          `json:"changes"`
	Config    *container.Config `json:"config"`
}

// WaitOptions holds parameters to wait for a container.
type WaitOptions struct {
	Condition string `query:"condition"`
}

// Containers is an interface that defines the methods that a containers implementation must provide.
type Containers interface {
	// List lists all containers running on the host.
	List(ctx context.Context, opts ListOptions) ([]types.Container, error)
	// Info returns detailed information about a container by ID.
	Info(ctx context.Context, id string) (*types.ContainerJSON, error)

	// Container lifecycle operations
	Start(ctx context.Context, id string, opts StartOptions) error
	Stop(ctx context.Context, id string, opts StopOptions) error
	Restart(ctx context.Context, id string, opts RestartOptions) error
	Kill(ctx context.Context, id string, opts KillOptions) error
	Remove(ctx context.Context, id string, opts RemoveOptions) error
	Pause(ctx context.Context, id string) error
	Unpause(ctx context.Context, id string) error

	// Container monitoring operations
	Logs(ctx context.Context, id string, opts LogsOptions) (io.ReadCloser, error)
	Stats(ctx context.Context, id string, opts StatsOptions) (container.StatsResponseReader, error)
	Top(ctx context.Context, id string, opts TopOptions) (container.ContainerTopOKBody, error)
	Changes(ctx context.Context, id string) ([]container.FilesystemChange, error)

	// Container file operations
	CopyToContainer(ctx context.Context, id string, path string, content io.Reader, opts CopyOptions) error
	CopyFromContainer(ctx context.Context, id string, path string) (io.ReadCloser, container.PathStat, error)

	// Container management operations
	Rename(ctx context.Context, id string, opts RenameOptions) error
	Update(ctx context.Context, id string, opts UpdateOptions) (container.ContainerUpdateOKBody, error)
	Resize(ctx context.Context, id string, opts ResizeOptions) error
	Commit(ctx context.Context, id string, opts CommitOptions) (types.IDResponse, error)
	Wait(ctx context.Context, id string, opts WaitOptions) (<-chan container.WaitResponse, <-chan error)
	Export(ctx context.Context, id string) (io.ReadCloser, error)
}
