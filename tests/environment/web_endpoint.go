package environment

import (
	"context"
	"crypto/tls"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/require"
)

// WebEndpointTLS is how a web endpoint reaches the service on the device: over TLS or not, whether
// it verifies the device's certificate, and the name it verifies it against and sends as Host.
type WebEndpointTLS struct {
	Enabled bool   `json:"enabled"`
	Verify  bool   `json:"verify"`
	Domain  string `json:"domain"`
}

// WebEndpointCreate is the body of POST /api/web-endpoints. TTL is in seconds, -1 for an endpoint
// that never expires.
type WebEndpointCreate struct {
	UID  string          `json:"uid"`
	Host string          `json:"host"`
	Port int             `json:"port"`
	TTL  int             `json:"ttl"`
	TLS  *WebEndpointTLS `json:"tls,omitempty"`
}

// WebEndpoint is a web endpoint as the enterprise API answers it. Address is the label the endpoint
// answers under, FullAddress that label under the instance's web endpoint domain. ExpiresIn is the
// zero time for an endpoint that never expires.
type WebEndpoint struct {
	Address     string         `json:"address"`
	FullAddress string         `json:"full_address"`
	Namespace   string         `json:"namespace"`
	DeviceUID   string         `json:"device_uid"`
	Host        string         `json:"host"`
	Port        int            `json:"port"`
	TTL         int            `json:"ttl"`
	TLS         WebEndpointTLS `json:"tls"`
	ExpiresIn   time.Time      `json:"expires_in"`
	CreatedAt   time.Time      `json:"created_at"`
}

// TunnelCreate is the body of the deprecated POST /api/devices/:uid/tunnels.
type TunnelCreate struct {
	Host string `json:"host"`
	Port int    `json:"port"`
	TTL  int    `json:"ttl"`
}

// Tunnel is a web endpoint as the deprecated device tunnel API lists it, naming its device's UID
// in Device.
type Tunnel struct {
	Address     string    `json:"address"`
	FullAddress string    `json:"full_address"`
	Namespace   string    `json:"namespace"`
	Device      string    `json:"device"`
	Host        string    `json:"host"`
	Port        int       `json:"port"`
	TTL         int       `json:"ttl"`
	ExpiresIn   time.Time `json:"expires_in"`
	CreatedAt   time.Time `json:"created_at"`
}

// PostWebEndpoint sends body to POST /api/web-endpoints as the bearer [DockerCompose.R] carries and
// returns the answer whatever its status code, decoding a 200 into result when result is not nil.
// ctx bounds the request. It returns the error of a request that never got an answer, or whose
// answer could not be read or decoded.
func (dc *DockerCompose) PostWebEndpoint(ctx context.Context, body any, result *WebEndpoint) (*resty.Response, error) {
	req := dc.R(ctx).SetBody(body)
	if result != nil {
		req = req.SetResult(result)
	}

	return req.Post("/api/web-endpoints")
}

// CreateWebEndpoint creates a web endpoint in the authenticated namespace and returns it as the
// server answered, failing t unless the server answers 200. It deletes the endpoint when t ends,
// ignoring the answer, so a test that already deleted it changes nothing.
func (dc *DockerCompose) CreateWebEndpoint(t *testing.T, req WebEndpointCreate) WebEndpoint {
	t.Helper()

	endpoint := WebEndpoint{}

	resp, err := dc.PostWebEndpoint(t.Context(), req, &endpoint)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

	t.Cleanup(func() {
		_, _ = dc.DeleteWebEndpoint(context.Background(), endpoint.Address)
	})

	return endpoint
}

// ListWebEndpoints returns the authenticated namespace's web endpoints, expired ones included, a
// single page of the maximum size, failing t unless the server answers 200.
func (dc *DockerCompose) ListWebEndpoints(t *testing.T) []WebEndpoint {
	t.Helper()

	endpoints := []WebEndpoint{}

	resp, err := dc.R(t.Context()).
		SetQueryParam("per_page", "100").
		SetResult(&endpoints).
		Get("/api/web-endpoints")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

	return endpoints
}

// DeleteWebEndpoint asks to remove the web endpoint address as the bearer [DockerCompose.R] carries
// and returns the answer whatever its status code. ctx bounds the request. It returns the error of
// a request that never got an answer or whose answer could not be read.
func (dc *DockerCompose) DeleteWebEndpoint(ctx context.Context, address string) (*resty.Response, error) {
	return dc.R(ctx).Delete("/api/web-endpoints/" + address)
}

// PostTunnel sends body to the deprecated POST /api/devices/:uid/tunnels as the bearer
// [DockerCompose.R] carries and returns the answer whatever its status code. ctx bounds the
// request. It returns the error of a request that never got an answer or whose answer could not be
// read.
func (dc *DockerCompose) PostTunnel(ctx context.Context, uid string, body TunnelCreate) (*resty.Response, error) {
	return dc.R(ctx).SetBody(body).Post("/api/devices/" + uid + "/tunnels")
}

// ListTunnels returns the web endpoints of the device uid through the deprecated tunnel API, a
// single page of the maximum size, failing t unless the server answers 200.
func (dc *DockerCompose) ListTunnels(t *testing.T, uid string) []Tunnel {
	t.Helper()

	tunnels := []Tunnel{}

	resp, err := dc.R(t.Context()).
		SetQueryParam("per_page", "100").
		SetResult(&tunnels).
		Get("/api/devices/" + uid + "/tunnels")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

	return tunnels
}

// DeleteTunnel asks the deprecated tunnel API to remove the web endpoint address of the device uid
// as the bearer [DockerCompose.R] carries and returns the answer whatever its status code. ctx
// bounds the request. It returns the error of a request that never got an answer or whose answer
// could not be read.
func (dc *DockerCompose) DeleteTunnel(ctx context.Context, uid, address string) (*resty.Response, error) {
	return dc.R(ctx).Delete("/api/devices/" + uid + "/tunnels/" + address)
}

// WebEndpointResponse is what the gateway answered a request sent to a web endpoint's address.
type WebEndpointResponse struct {
	StatusCode int
	Body       string
}

// VisitWebEndpoint sends GET uri to the gateway's HTTP port under the Host host, as a browser that
// resolved a web endpoint's address to the gateway does, carrying no credential, on a connection
// of its own that it closes once the answer is read. ctx bounds the request. It returns the error
// of a uri that does not form a URL, of a request that never got an answer, and of an answer whose
// body could not be read in full or closed.
func (dc *DockerCompose) VisitWebEndpoint(ctx context.Context, host, uri string) (*WebEndpointResponse, error) {
	return visit(ctx, &http.Transport{DisableKeepAlives: true}, dc.BaseURL()+uri, host)
}

// VisitWebEndpointOverTLS sends GET uri to the gateway's HTTPS port the way
// [DockerCompose.VisitWebEndpoint] does to its HTTP port, naming host in TLS's server name
// indication as well as in Host. config verifies the gateway's certificate; its ServerName is
// replaced with host. It returns the errors [DockerCompose.VisitWebEndpoint] does, among them a
// certificate config refuses.
func (dc *DockerCompose) VisitWebEndpointOverTLS(ctx context.Context, config *tls.Config, host, uri string) (*WebEndpointResponse, error) {
	config = config.Clone()
	config.ServerName = host

	return visit(ctx, &http.Transport{DisableKeepAlives: true, TLSClientConfig: config}, "https://"+dc.HTTPSAddress()+uri, host)
}

func visit(ctx context.Context, transport *http.Transport, url, host string) (*WebEndpointResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Host = host

	resp, err := (&http.Client{Transport: transport}).Do(req)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err := errors.Join(err, resp.Body.Close()); err != nil {
		return nil, err
	}

	return &WebEndpointResponse{StatusCode: resp.StatusCode, Body: string(body)}, nil
}
