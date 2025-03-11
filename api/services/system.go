package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/shellhub-io/shellhub/api/pkg/responses"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/envs"
)

type SystemService interface {
	// GetSystemInfo retrieves the instance's information
	GetSystemInfo(ctx context.Context, req *requests.GetSystemInfo) (*responses.SystemInfo, error)

	SystemDownloadInstallScript(ctx context.Context) (string, error)
}

func (s *service) GetSystemInfo(ctx context.Context, req *requests.GetSystemInfo) (*responses.SystemInfo, error) {
	apiHost := strings.Split(req.Host, ":")[0]
	sshPort := envs.DefaultBackend.Get("SHELLHUB_SSH_PORT")

	resp := &responses.SystemInfo{
		Version: envs.DefaultBackend.Get("SHELLHUB_VERSION"),
		Setup:   true,
		Endpoints: &responses.SystemEndpointsInfo{
			API: apiHost,
			SSH: fmt.Sprintf("%s:%s", apiHost, sshPort),
		},
	}

	if req.Port > 0 {
		resp.Endpoints.API = fmt.Sprintf("%s:%d", apiHost, req.Port)
	} else {
		resp.Endpoints.API = req.Host
	}

	return resp, nil
}

func (s *service) SystemDownloadInstallScript(_ context.Context) (string, error) {
	return "", nil
}
