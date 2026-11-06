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
	"github.com/shellhub-io/shellhub/server/api/pkg/origin"
	"github.com/shellhub-io/shellhub/server/api/pkg/responses"
	log "github.com/sirupsen/logrus"
)

// SystemService reports what the instance is and how to join it — the only service with
// operations reachable before authentication.
type SystemService interface {
	// GetSystemInfo retrieves the instance's information, with the API and SSH endpoints at the
	// host and port reached names. It returns the store's error when the system cannot be read.
	GetSystemInfo(ctx context.Context, reached origin.Origin) (*responses.SystemInfo, error)

	// SystemDownloadInstallScript renders the agent install script, injecting
	// instance-derived defaults (notably SERVER_ADDRESS from the scheme, host and port reached
	// names) so the user does not have to pass them on the command line. It returns the os.ReadFile
	// error when the install.sh template cannot be read.
	SystemDownloadInstallScript(ctx context.Context, req *requests.SystemInstallScript, reached origin.Origin) (string, error)
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

func (s *service) GetSystemInfo(ctx context.Context, reached origin.Origin) (*responses.SystemInfo, error) {
	system, err := s.systemGet(ctx)
	if err != nil {
		return nil, err
	}

	apiHost, _, _ := strings.Cut(reached.Host, ":")
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

	if reached.Port != "" {
		resp.Endpoints.API = apiHost + ":" + reached.Port
	} else {
		resp.Endpoints.API = reached.Host
	}

	return resp, nil
}

func (s *service) SystemDownloadInstallScript(_ context.Context, req *requests.SystemInstallScript, reached origin.Origin) (string, error) {
	raw, err := os.ReadFile("/templates/install.sh")
	if err != nil {
		return "", err
	}

	return renderInstallScript(string(raw), req, reached), nil
}

func renderInstallScript(template string, req *requests.SystemInstallScript, reached origin.Origin) string {
	return strings.Replace(template, "{{.Overrides}}", buildInstallOverrides(req, reached), 1)
}

func buildInstallOverrides(req *requests.SystemInstallScript, reached origin.Origin) string {
	scheme := reached.Scheme
	if scheme != "http" && scheme != "https" {
		scheme = "https"
	}

	var b strings.Builder
	b.WriteString("\n")

	if address := ClientFacingAddress(scheme, reached.Host, reached.Port); address != "" {
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
// on, for an absolute URL handed back to it. The port is port, the one the client connected to as
// [origin.Origin] resolves it, or else the port carried on host when port is empty. It is left out
// when it is scheme's default. An IPv6 host comes back in brackets, and an empty host returns an
// empty string.
func ClientFacingAddress(scheme, host, port string) string {
	if host == "" {
		return ""
	}

	name, hostPort := host, ""
	if h, p, err := net.SplitHostPort(host); err == nil {
		name, hostPort = h, p
	}

	if port == "" {
		port = hostPort
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
