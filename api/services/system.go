package services

import (
	"context"
	"fmt"
	"os"
	"strings"
	"text/template"

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

	tmpl, err := template.New("install").Parse(string(raw))
	if err != nil {
		return "", err
	}

	var out strings.Builder
	if err := tmpl.Execute(&out, map[string]string{"Overrides": buildInstallOverrides(req)}); err != nil {
		return "", err
	}

	return out.String(), nil
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

	// Host and port come from forwarded headers (nginx $host drops the port, so
	// it arrives separately). Sanitize before embedding in a shell string, and
	// only append the port when it is non-standard for the scheme.
	if host := sanitizeInstallHost(strings.Split(req.Host, ":")[0]); host != "" {
		address := host
		if port := sanitizeInstallHost(req.ForwardedPort); port != "" && !isDefaultPort(scheme, port) {
			address = host + ":" + port
		}

		fmt.Fprintf(&b, "SERVER_ADDRESS=\"${SERVER_ADDRESS:-%s://%s}\"\n", scheme, address)
	}

	if req.TenantID != "" {
		fmt.Fprintf(&b, "TENANT_ID=\"${TENANT_ID:-%s}\"\n", sanitizeInstallHost(req.TenantID))
	}

	if req.PreferredHostname != "" {
		fmt.Fprintf(&b, "PREFERRED_HOSTNAME=\"${PREFERRED_HOSTNAME:-%s}\"\n", sanitizeInstallHost(req.PreferredHostname))
	}

	if req.PreferredIdentity != "" {
		fmt.Fprintf(&b, "PREFERRED_IDENTITY=\"${PREFERRED_IDENTITY:-%s}\"\n", sanitizeInstallHost(req.PreferredIdentity))
	}

	return b.String()
}

// isDefaultPort reports whether port is the default for the scheme, in which
// case it should be omitted from the server address.
func isDefaultPort(scheme, port string) bool {
	return (scheme == "https" && port == "443") || (scheme == "http" && port == "80")
}

// sanitizeInstallHost keeps only characters valid in a host[:port], tenant UUID
// or hostname/identity, dropping anything that could break out of the shell
// double-quoted string the value is embedded in.
func sanitizeInstallHost(value string) string {
	return strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
			return r
		case r == '.' || r == '-' || r == ':' || r == '_':
			return r
		default:
			return -1
		}
	}, value)
}
