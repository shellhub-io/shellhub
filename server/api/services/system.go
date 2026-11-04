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
		cached.Setup && cached.InstanceID != "" && cached.Authentication != nil && cached.Authentication.Local != nil {

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

	apiHost, _, _ := strings.Cut(req.Host, ":")
	resp := &responses.SystemInfo{
		Version: s.version,
		Setup:   system.Setup,
		Endpoints: &responses.SystemEndpointsInfo{
			API: apiHost,
			SSH: fmt.Sprintf("%s:%s", apiHost, s.sshPort),
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

	return renderInstallScript(string(raw), req), nil
}

func renderInstallScript(template string, req *requests.SystemInstallScript) string {
	return strings.Replace(template, "{{.Overrides}}", buildInstallOverrides(req), 1)
}

func buildInstallOverrides(req *requests.SystemInstallScript) string {
	scheme := req.Scheme
	if scheme != "http" && scheme != "https" {
		scheme = "https"
	}

	var b strings.Builder
	b.WriteString("\n")

	if address := ClientFacingAddress(scheme, req.Host, req.ForwardedPort); address != "" {
		writeInstallOverride(&b, "SERVER_ADDRESS", scheme+"://"+address)
	}

	writeInstallOverride(&b, "TENANT_ID", req.TenantID)
	writeInstallOverride(&b, "PREFERRED_HOSTNAME", req.PreferredHostname)
	writeInstallOverride(&b, "PREFERRED_IDENTITY", req.PreferredIdentity)

	return b.String()
}

func writeInstallOverride(b *strings.Builder, name, value string) {
	if value == "" {
		return
	}

	fmt.Fprintf(b, "[ -n \"${%s:-}\" ] || %s=%s\n", name, name, singleQuote(value))
}

func singleQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", `'\''`) + "'"
}

// ClientFacingAddress returns the host, and port when it matters, that a client reached the server
// on, for an absolute URL handed back to it. The port is forwardedPort, the X-Forwarded-Port the
// gateway sets because it forwards Host without one, or else the port carried on host. It is left
// out when it is scheme's default. An IPv6 host comes back in brackets, and an empty host returns
// an empty string.
func ClientFacingAddress(scheme, host, forwardedPort string) string {
	if host == "" {
		return ""
	}

	name, port := host, ""
	if h, p, err := net.SplitHostPort(host); err == nil {
		name, port = h, p
	}

	if forwardedPort != "" {
		port = forwardedPort
	}

	name = strings.Trim(name, "[]")

	defaultPort := (scheme == "https" && port == "443") || (scheme == "http" && port == "80")

	switch {
	case port != "" && !defaultPort:
		return net.JoinHostPort(name, port)
	case strings.Contains(name, ":"):
		return "[" + name + "]"
	default:
		return name
	}
}
