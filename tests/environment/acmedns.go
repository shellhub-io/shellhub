package environment

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/moby/moby/api/types/container"
	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// ACMEDNSWebEndpointsDomain is the web endpoint domain of a stack started with [Config.ACMEDNS]:
// the domain whose wildcard the gateway proves over DNS-01 through the stack's acme-dns server.
const ACMEDNSWebEndpointsDomain = "endpoints." + acmeDNSZone

const (
	acmeDNSZone    = "example.test"
	acmeDNSImage   = "docker.io/joohoi/acme-dns:v2.0.2"
	acmeDNSAPIPort = "8053"

	acmeDNSDirEnv   = "SHELLHUB_TEST_ACME_DNS_DIR"
	acmeDNSImageEnv = "SHELLHUB_TEST_ACME_DNS_IMAGE"
	acmeDNSZoneEnv  = "SHELLHUB_TEST_ACME_DNS_ZONE"

	acmeDNSDirName    = "acme-dns"
	acmeDNSConfigFile = "config.cfg"
	acmeDNSDataDir    = "data"
)

var errACMEDNSWithoutACME = errors.New("acme-dns answers the gateway's DNS-01 challenge to the ACME server, so it needs ACME")

type acmeDNSAccount struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	FullDomain string `json:"fulldomain"`
	Subdomain  string `json:"subdomain"`
}

func (cfg Config) acmeDNSLayer(ctx context.Context, acmeDir string) (httpsLayer, error) {
	dir := filepath.Join(acmeDir, acmeDNSDirName)
	if err := os.MkdirAll(filepath.Join(dir, acmeDNSDataDir), 0o750); err != nil {
		return httpsLayer{}, err
	}

	if err := writeACMEDNSConfig(dir); err != nil {
		return httpsLayer{}, err
	}

	account, err := registerACMEDNSAccount(ctx, cfg.Run, dir)
	if err != nil {
		return httpsLayer{}, err
	}

	if err := writeACMEDNSConfig(dir,
		"ns."+acmeDNSZone+". A 127.0.0.1",
		acmeDNSZone+". NS ns."+acmeDNSZone+".",
		"_acme-challenge."+ACMEDNSWebEndpointsDomain+". CNAME "+account.FullDomain+".",
	); err != nil {
		return httpsLayer{}, err
	}

	envs := map[string]string{
		acmeDNSDirEnv:                               dir,
		acmeDNSImageEnv:                             acmeDNSImage,
		acmeDNSZoneEnv:                              acmeDNSZone,
		"SHELLHUB_WEB_ENDPOINTS_DOMAIN":             ACMEDNSWebEndpointsDomain,
		"SHELLHUB_WEB_ENDPOINTS_DNS_PROVIDER":       "acmedns",
		"SHELLHUB_WEB_ENDPOINTS_ACME_DNS_URL":       "http://127.0.0.1:" + acmeDNSAPIPort,
		"SHELLHUB_WEB_ENDPOINTS_ACME_DNS_USERNAME":  account.Username,
		"SHELLHUB_WEB_ENDPOINTS_ACME_DNS_PASSWORD":  account.Password,
		"SHELLHUB_WEB_ENDPOINTS_ACME_DNS_SUBDOMAIN": account.Subdomain,
	}

	return httpsLayer{envs: envs, composeFiles: []string{"../docker-compose.acmedns.test.yml"}}, nil
}

func writeACMEDNSConfig(dir string, records ...string) error {
	quoted := make([]string, len(records))
	for i, record := range records {
		quoted[i] = fmt.Sprintf("%q", record)
	}

	config := fmt.Sprintf(`[general]
listen = "0.0.0.0:53"
protocol = "both4"
domain = %[1]q
nsname = "ns.%[1]s"
nsadmin = "admin.%[1]s"
records = [%[2]s]

[database]
engine = "sqlite"
connection = "/var/lib/acme-dns/acme-dns.db"

[api]
ip = "0.0.0.0"
port = %[3]q
tls = "none"
disable_registration = false

[logconfig]
loglevel = "debug"
logtype = "stdout"
logformat = "text"
`, acmeDNSZone, strings.Join(quoted, ", "), acmeDNSAPIPort)

	return os.WriteFile(filepath.Join(dir, acmeDNSConfigFile), []byte(config), 0o600)
}

func registerACMEDNSAccount(ctx context.Context, run *Run, dir string) (acmeDNSAccount, error) {
	server, err := tc.GenericContainer(ctx, tc.GenericContainerRequest{
		ContainerRequest: tc.ContainerRequest{
			Image:  acmeDNSImage,
			Labels: run.Labels(),
			HostConfigModifier: func(hc *container.HostConfig) {
				hc.NetworkMode = "none"
				hc.Binds = []string{
					filepath.Join(dir, acmeDNSConfigFile) + ":/etc/acme-dns/config.cfg:ro",
					filepath.Join(dir, acmeDNSDataDir) + ":/var/lib/acme-dns",
				}
			},
			WaitingFor: wait.ForLog("Listening HTTP"),
		},
		Started: true,
		Logger:  log.New(io.Discard, "", log.LstdFlags),
	})
	if server != nil {
		defer server.Terminate(context.WithoutCancel(ctx)) //nolint:errcheck // the account is in the data directory; a container left behind is the run's to sweep
	}

	if err != nil {
		return acmeDNSAccount{}, fmt.Errorf("start acme-dns to register an account: %w", err)
	}

	docker, ok := server.(*tc.DockerContainer)
	if !ok {
		return acmeDNSAccount{}, fmt.Errorf("acme-dns started as a %T, not a Docker container", server)
	}

	body, err := execOutput(ctx, docker, []string{"wget", "-qO-", "--post-data=", "http://127.0.0.1:" + acmeDNSAPIPort + "/register"})
	if err != nil {
		return acmeDNSAccount{}, fmt.Errorf("register an acme-dns account: %w", err)
	}

	var account acmeDNSAccount
	if err := json.Unmarshal(body, &account); err != nil {
		return acmeDNSAccount{}, fmt.Errorf("read the acme-dns account %q: %w", body, err)
	}

	return account, nil
}
