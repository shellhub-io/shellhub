package services

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/shellhub-io/shellhub/api/pkg/responses"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/envs"
)

type SystemService interface {
	// GetSystemInfo retrieves the instance's information
	GetSystemInfo(ctx context.Context, req *requests.GetSystemInfo) (*responses.SystemInfo, error)

	// SystemDownloadInstallScript renders the agent install script, injecting
	// instance-derived defaults (notably SERVER_ADDRESS from the request host)
	// so the user does not have to pass them on the command line.
	SystemDownloadInstallScript(ctx context.Context, req *requests.SystemInstallScript) (string, error)
}

func (s *service) GetSystemInfo(ctx context.Context, req *requests.GetSystemInfo) (*responses.SystemInfo, error) {
	system, err := s.store.SystemGet(ctx)
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

	// Replace only the marker, not through text/template: the script contains
	// other "{{...}}" sequences (Docker/Podman --format '{{.Names}}') that the
	// template engine would resolve to "<no value>" and silently break.
	overrides := buildInstallOverrides(req)

	return strings.Replace(string(raw), "{{.Overrides}}", overrides, 1), nil
}

// buildInstallOverrides renders the shell block injected at the {{.Overrides}}
// marker. The marker sits on a comment line, so the block starts with a newline
// to break onto real assignment lines. Each value is a default ("${VAR:-...}")
// so an explicit env the user passes still wins.
func buildInstallOverrides(req *requests.SystemInstallScript) string {
	scheme := req.Scheme
	if scheme != "http" && scheme != "https" {
		scheme = "https"
	}

	var b strings.Builder
	b.WriteString("\n")

	// The host may carry a port: the gateway forwards it separately in
	// X-Forwarded-Port, but the direct-access fallback (c.Request().Host) keeps
	// it inline. Split it out so it isn't lost, preferring the forwarded port.
	// Values are reflected from the requester's own request and the script is
	// served uncached (Cache-Control: no-store), so it only ever reaches that
	// requester; no escaping is needed.
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

// isDefaultPort reports whether port is the default for the scheme, in which
// case it should be omitted from the server address.
func isDefaultPort(scheme, port string) bool {
	return (scheme == "https" && port == "443") || (scheme == "http" && port == "80")
}
