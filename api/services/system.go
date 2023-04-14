package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/models"
)

type SystemService interface {
	SystemGetInfo(ctx context.Context, host string, port int) (*models.SystemInfo, error)
	SystemDownloadInstallScript(ctx context.Context) error
}

// SystemGetInfo returns system instance information.
// It receives a context (ctx), a host (host) which is used to determine the
// API and SSH host of the system, and a port (port) that can be specified
// to override the API port from the host.
func (s *service) SystemGetInfo(ctx context.Context, host string, port int) (*models.SystemInfo, error) {
	apiHost := strings.Split(host, ":")[0]
	sshPort := envs.DefaultBackend.Get("SHELLHUB_SSH_PORT")

	info := &models.SystemInfo{
		Version: envs.DefaultBackend.Get("SHELLHUB_VERSION"),
		Endpoints: &models.SystemInfoEndpoints{
			API: apiHost,
			SSH: fmt.Sprintf("%s:%s", apiHost, sshPort),
		},
	}

	if port > 0 {
		info.Endpoints.API = fmt.Sprintf("%s:%d", apiHost, port)
	} else {
		info.Endpoints.API = host
	}

	return info, nil
}

func (s *service) SystemDownloadInstallScript(ctx context.Context) error {
	return nil
}
