package services

import (
	"context"
	"fmt"
	"strings"
	"text/template"

	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/models"
)

type SystemService interface {
	SystemGetInfo(ctx context.Context, req requests.SystemGetInfo) (*models.SystemInfo, error)
	SystemDownloadInstallScript(ctx context.Context, req requests.SystemInstallScript) (*template.Template, map[string]interface{}, error)
}

// SystemGetInfo returns system instance information.
// It receives a context (ctx) and requests.SystemGetInfo, what contains a host (host) which is used to determine the
// API and SSH host of the system, and a port (port) that can be specified to override the API port from the host.
func (s *service) SystemGetInfo(_ context.Context, req requests.SystemGetInfo) (*models.SystemInfo, error) {
	apiHost := strings.Split(req.Host, ":")[0]
	sshPort := envs.DefaultBackend.Get("SHELLHUB_SSH_PORT")

	info := &models.SystemInfo{
		Version: envs.DefaultBackend.Get("SHELLHUB_VERSION"),
		Endpoints: &models.SystemInfoEndpoints{
			API: apiHost,
			SSH: fmt.Sprintf("%s:%s", apiHost, sshPort),
		},
	}

	if req.Port > 0 {
		info.Endpoints.API = fmt.Sprintf("%s:%d", apiHost, req.Port)
	} else {
		info.Endpoints.API = req.Host
	}

	return info, nil
}

func (s *service) SystemDownloadInstallScript(_ context.Context, req requests.SystemInstallScript) (*template.Template, map[string]interface{}, error) {
	tmpl, err := template.ParseFiles("./templates/kickstart.sh")
	if err != nil {
		return nil, nil, err
	}

	return tmpl, map[string]interface{}{
		"scheme":             req.Scheme,
		"host":               req.Host,
		"tenant_id":          req.TenantID,
		"keepalive_interval": req.KeepAliveInternavel,
		"preferred_hostname": req.PreferredHostname,
		"preferred_identity": req.PreferredIdentity,
		"version":            envs.DefaultBackend.Get("SHELLHUB_VERSION"),
	}, nil
}
