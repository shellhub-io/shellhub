package services

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/pkg/responses"
	log "github.com/sirupsen/logrus"
)

// SystemService reports what the instance is and how to join it — the only service with
// operations reachable before authentication.
type SystemService interface {
	// GetSystemInfo retrieves the instance's information
	GetSystemInfo(ctx context.Context, req *requests.GetSystemInfo) (*responses.SystemInfo, error)

	// SystemDownloadInstallScript renders the agent install script, injecting
	// instance-derived defaults (notably SERVER_ADDRESS from the request host)
	// so the user does not have to pass them on the command line.
	SystemDownloadInstallScript(ctx context.Context, req *requests.SystemInstallScript) (string, error)
}

const systemCacheTTL = time.Minute

func (s *service) systemGet(ctx context.Context) (*models.System, error) {
	cached := new(models.System)
	if err := s.cache.Get(ctx, cache.SystemKey, cached); err == nil &&
		cached.Setup && cached.Authentication != nil && cached.Authentication.Local != nil {

		return cached, nil
	}

	system, err := s.store.SystemGet(ctx)
	if err != nil {
		return nil, err
	}

	if system.Setup {
		if err := s.cache.Set(ctx, cache.SystemKey, system, systemCacheTTL); err != nil {
			log.WithError(err).Warn("failed to cache the system row")
		}
	}

	return system, nil
}

func (s *service) GetSystemInfo(ctx context.Context, req *requests.GetSystemInfo) (*responses.SystemInfo, error) {
	system, err := s.systemGet(ctx)
	if err != nil {
		return nil, err
	}

	apiHost := strings.Split(req.Host, ":")[0]
	sshPort := envs.DefaultBackend.Get("SHELLHUB_SSH_PORT")

	resp := &responses.SystemInfo{
		Version: envs.DefaultBackend.Get("SHELLHUB_VERSION"),
		Setup:   system.Setup,
		Endpoints: &responses.SystemEndpointsInfo{
			API: apiHost,
			SSH: fmt.Sprintf("%s:%s", apiHost, sshPort),
		},
		Authentication: &responses.SystemAuthenticationInfo{
			Local: system.Authentication.Local.Enabled,
		},
	}

	if req.Port > 0 {
		resp.Endpoints.API = fmt.Sprintf("%s:%d", apiHost, req.Port)
	} else {
		resp.Endpoints.API = req.Host
	}

	return resp, nil
}

func (s *service) SystemDownloadInstallScript(_ context.Context, req *requests.SystemInstallScript) (string, error) {
	raw, err := os.ReadFile("/templates/install.sh")
	if err != nil {
		return "", err
	}

	overrides := buildInstallOverrides(req)

	return strings.Replace(string(raw), "{{.Overrides}}", overrides, 1), nil
}

func buildInstallOverrides(req *requests.SystemInstallScript) string {
	scheme := req.Scheme
	if scheme != "http" && scheme != "https" {
		scheme = "https"
	}

	var b strings.Builder
	b.WriteString("\n")

	host, hostPort := req.Host, ""
	if h, p, err := net.SplitHostPort(req.Host); err == nil {
		host, hostPort = h, p
	}

	if host != "" {
		port := req.ForwardedPort
		if port == "" {
			port = hostPort
		}

		address := host
		if port != "" && !isDefaultPort(scheme, port) {
			address = net.JoinHostPort(host, port)
		}

		fmt.Fprintf(&b, "SERVER_ADDRESS=\"${SERVER_ADDRESS:-%s://%s}\"\n", scheme, address)
	}

	if req.TenantID != "" {
		fmt.Fprintf(&b, "TENANT_ID=\"${TENANT_ID:-%s}\"\n", req.TenantID)
	}

	if req.PreferredHostname != "" {
		fmt.Fprintf(&b, "PREFERRED_HOSTNAME=\"${PREFERRED_HOSTNAME:-%s}\"\n", req.PreferredHostname)
	}

	if req.PreferredIdentity != "" {
		fmt.Fprintf(&b, "PREFERRED_IDENTITY=\"${PREFERRED_IDENTITY:-%s}\"\n", req.PreferredIdentity)
	}

	return b.String()
}

func isDefaultPort(scheme, port string) bool {
	return (scheme == "https" && port == "443") || (scheme == "http" && port == "80")
}
