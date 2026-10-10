package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/moby/moby/api/types/container"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/testimage"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	deviceTLSDomain        = "device.example.com"
	webEndpointDomain      = "localhost"
	unknownAddress         = "0123456789abcdef0123456789abcdef"
	webEndpointNotFound    = "web endpoint not found"
	gatewayRootCertificate = "/data/pki/authorities/local/root.crt"
	shortTTL               = 3
)

var webEndpointAddress = regexp.MustCompile(`^[0-9a-f]{32}$`)

type deviceService struct {
	name     string
	httpPort int
	tlsPort  int
}

type deviceSeen struct {
	Name string `json:"name"`
	Host string `json:"host"`
	URI  string `json:"uri"`
	TLS  bool   `json:"tls"`
}

// TestEnterpriseWebEndpoints covers web endpoints on an enterprise instance that enables them: what
// the API refuses to create, the address and lifetime of what it creates, the gateway carrying a
// request sent to an endpoint's address through the agent into an HTTP or TLS service on the
// device, the deprecated tunnel API over the same endpoints, and a device merge handing a device's
// endpoints to the device it merged into. The device's services listen in the agents' network
// namespace and answer with the name they were started with and what they received.
func TestEnterpriseWebEndpoints(t *testing.T) {
	ctx := context.Background()

	compose := newEnterpriseEnvironment(t, ctx, environment.New(t, run).WithEnv("SHELLHUB_WEB_ENDPOINTS", "true"))
	_, device := startAcceptedAgent(t, ctx, compose)

	first := startDeviceService(t, compose, deviceService{name: "first", httpPort: 8001, tlsPort: 8443})
	second := startDeviceService(t, compose, deviceService{name: "second", httpPort: 8002, tlsPort: 8444})

	plain := func(service deviceService, ttl int) environment.WebEndpointCreate {
		return environment.WebEndpointCreate{UID: device.UID, Host: "127.0.0.1", Port: service.httpPort, TTL: ttl}
	}

	t.Run("creating", func(t *testing.T) { testWebEndpointCreation(t, compose, device, plain(first, -1)) })
	t.Run("proxying", func(t *testing.T) { testWebEndpointProxy(t, compose, device, first, second) })
	t.Run("lifetime", func(t *testing.T) { testWebEndpointLifetime(t, compose, plain(first, -1), plain(second, -1)) })
	t.Run("deprecated tunnel API", func(t *testing.T) { testTunnelAPI(t, compose, device, second) })
	t.Run("device merge", func(t *testing.T) { testWebEndpointDeviceMerge(t, ctx, compose, first) })
}

// TestEnterpriseWebEndpointsDisabled covers an enterprise instance that leaves web endpoints off,
// as it does by default: neither the API nor the gateway serves them.
func TestEnterpriseWebEndpointsDisabled(t *testing.T) {
	compose := newEnterpriseEnvironment(t, context.Background(), environment.New(t, run))

	t.Run("the API serves no web endpoint route", func(t *testing.T) {
		requests := []struct {
			method string
			path   string
		}{
			{method: http.MethodGet, path: "/api/web-endpoints"},
			{method: http.MethodPost, path: "/api/web-endpoints"},
			{method: http.MethodDelete, path: "/api/web-endpoints/" + unknownAddress},
			{method: http.MethodGet, path: "/api/devices/" + strings.Repeat("0", 64) + "/tunnels"},
			{method: http.MethodPost, path: "/api/devices/" + strings.Repeat("0", 64) + "/tunnels"},
		}

		for _, req := range requests {
			resp, err := compose.R(t.Context()).SetBody(map[string]any{}).Execute(req.method, req.path)
			require.NoError(t, err)
			assert.Equal(t, http.StatusNotFound, resp.StatusCode(), "%s %s: %s", req.method, req.path, resp.String())
		}
	})

	t.Run("the gateway serves no site for an endpoint's address", func(t *testing.T) {
		resp := visitWebEndpoint(t, compose, unknownAddress+"."+webEndpointDomain, "/")

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		assert.Empty(t, resp.Body)
	})
}

// TestEnterpriseWebEndpointsTLS covers the gateway serving web endpoints over HTTPS on a domain no
// public authority signs for: it presents a certificate its own authority issued for the address,
// and carries the request to the server.
func TestEnterpriseWebEndpointsTLS(t *testing.T) {
	compose := environment.New(t, run).
		WithEdition(environment.EditionEnterprise).
		WithEnv("SHELLHUB_WEB_ENDPOINTS", "true").
		WithAutoSSL().
		Up(context.Background())
	t.Cleanup(compose.Down)

	roots := x509.NewCertPool()
	roots.AddCert(gatewayAuthority(t, compose))

	var resp *environment.WebEndpointResponse

	require.EventuallyWithT(t, func(tt *assert.CollectT) {
		var err error

		resp, err = compose.VisitWebEndpointOverTLS(t.Context(), &tls.Config{RootCAs: roots, MinVersion: tls.VersionTLS12}, unknownAddress+"."+webEndpointDomain, "/")
		assert.NoError(tt, err)
	}, 30*time.Second, time.Second)

	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	assert.Contains(t, resp.Body, webEndpointNotFound)
}

func testWebEndpointCreation(t *testing.T, compose *environment.DockerCompose, device *models.Device, valid environment.WebEndpointCreate) {
	t.Helper()

	refusals := []struct {
		description string
		change      func(req *environment.WebEndpointCreate)
		status      int
	}{
		{description: "a host that is not an IP address", change: func(req *environment.WebEndpointCreate) { req.Host = "localhost" }, status: http.StatusBadRequest},
		{description: "port 0", change: func(req *environment.WebEndpointCreate) { req.Port = 0 }, status: http.StatusBadRequest},
		{description: "a TTL of 0", change: func(req *environment.WebEndpointCreate) { req.TTL = 0 }, status: http.StatusBadRequest},
		{description: "a TTL below -1", change: func(req *environment.WebEndpointCreate) { req.TTL = -2 }, status: http.StatusBadRequest},
		{description: "a device the namespace does not hold", change: func(req *environment.WebEndpointCreate) { req.UID = strings.Repeat("0", 64) }, status: http.StatusNotFound},
		{description: "TLS without a domain", change: func(req *environment.WebEndpointCreate) { req.TLS = &environment.WebEndpointTLS{Enabled: true} }, status: http.StatusBadRequest},
	}

	for _, tc := range refusals {
		t.Run("an endpoint for "+tc.description+" is refused", func(t *testing.T) {
			req := valid
			tc.change(&req)

			requireCreateRefused(t, compose, req, tc.status)
		})
	}

	t.Run("an endpoint for a device not yet accepted is refused", func(t *testing.T) {
		pending := enrollPendingDevice(t, compose, "web-endpoint-pending", "02:00:00:00:22:01")

		req := valid
		req.UID = pending.UID

		requireCreateRefused(t, compose, req, http.StatusForbidden)
	})

	t.Run("an endpoint for a namespace that no longer exists is refused", func(t *testing.T) {
		const username, password = "orphan", "password"

		compose.NewUser(t, username, "orphan@shellhub.example.com", password)
		compose.NewNamespace(t, username, "orphanspace", "00000000-0000-4000-0000-000000002201", "")
		token := compose.AuthUser(t, username, password).Token

		compose.DeleteNamespace(t, "orphanspace")

		resp, err := compose.Anonymous(t.Context()).SetAuthToken(token).SetBody(valid).Post("/api/web-endpoints")
		require.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode(), resp.String())
	})

	t.Run("an endpoint answers under a 32 hex digit address of the web endpoint domain", func(t *testing.T) {
		endpoint := compose.CreateWebEndpoint(t, valid)

		assert.Regexp(t, webEndpointAddress, endpoint.Address)
		assert.Equal(t, endpoint.Address+"."+webEndpointDomain, endpoint.FullAddress)
		assert.Equal(t, ShellHubNamespace, endpoint.Namespace)
		assert.Equal(t, device.UID, endpoint.DeviceUID)
		assert.Equal(t, -1, endpoint.TTL)
		assert.True(t, endpoint.ExpiresIn.IsZero(), "an endpoint with a TTL of -1 has no expiry")
	})

	t.Run("the same device, host and port keep their endpoint and get a new address once it is deleted", func(t *testing.T) {
		endpoint := compose.CreateWebEndpoint(t, valid)

		assert.Equal(t, endpoint.Address, compose.CreateWebEndpoint(t, valid).Address)

		resp, err := compose.DeleteWebEndpoint(t.Context(), endpoint.Address)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

		recreated := compose.CreateWebEndpoint(t, valid)
		assert.Regexp(t, webEndpointAddress, recreated.Address)
		assert.NotEqual(t, endpoint.Address, recreated.Address)
	})
}

func testWebEndpointProxy(t *testing.T, compose *environment.DockerCompose, device *models.Device, first, second deviceService) {
	t.Helper()

	t.Run("a request to an endpoint's address reaches the device's service with its path and the address as Host", func(t *testing.T) {
		endpoint := compose.CreateWebEndpoint(t, environment.WebEndpointCreate{UID: device.UID, Host: "127.0.0.1", Port: first.httpPort, TTL: -1})

		seen := requireDeviceAnswer(t, compose, endpoint.FullAddress, "/some/path?query=value")
		assert.Equal(t, first.name, seen.Name)
		assert.Equal(t, "/some/path?query=value", seen.URI)
		assert.Equal(t, endpoint.FullAddress, seen.Host)
		assert.False(t, seen.TLS)
	})

	t.Run("requests to two endpoints each reach their own service", func(t *testing.T) {
		one := compose.CreateWebEndpoint(t, environment.WebEndpointCreate{UID: device.UID, Host: "127.0.0.1", Port: first.httpPort, TTL: -1})
		two := compose.CreateWebEndpoint(t, environment.WebEndpointCreate{UID: device.UID, Host: "127.0.0.1", Port: second.httpPort, TTL: -1})

		for range 3 {
			assert.Equal(t, first.name, requireDeviceAnswer(t, compose, one.FullAddress, "/").Name)
			assert.Equal(t, second.name, requireDeviceAnswer(t, compose, two.FullAddress, "/").Name)
		}
	})

	t.Run("a request to a deleted endpoint's address is refused", func(t *testing.T) {
		endpoint := compose.CreateWebEndpoint(t, environment.WebEndpointCreate{UID: device.UID, Host: "127.0.0.1", Port: first.httpPort, TTL: -1})
		requireDeviceAnswer(t, compose, endpoint.FullAddress, "/")

		resp, err := compose.DeleteWebEndpoint(t.Context(), endpoint.Address)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

		requireEndpointRefused(t, compose, endpoint.FullAddress)
	})

	t.Run("a request to an address no endpoint has is refused", func(t *testing.T) {
		requireEndpointRefused(t, compose, unknownAddress+"."+webEndpointDomain)
	})

	t.Run("an endpoint over TLS that does not verify reaches a self-signed service, naming its domain as Host", func(t *testing.T) {
		endpoint := compose.CreateWebEndpoint(t, environment.WebEndpointCreate{
			UID:  device.UID,
			Host: "127.0.0.1",
			Port: first.tlsPort,
			TTL:  -1,
			TLS:  &environment.WebEndpointTLS{Enabled: true, Verify: false, Domain: deviceTLSDomain},
		})

		seen := requireDeviceAnswer(t, compose, endpoint.FullAddress, "/secure?query=value")
		assert.Equal(t, first.name, seen.Name)
		assert.True(t, seen.TLS)
		assert.Equal(t, "/secure?query=value", seen.URI)
		assert.Equal(t, deviceTLSDomain, seen.Host)
	})

	t.Run("an endpoint over TLS that verifies refuses a self-signed service", func(t *testing.T) {
		endpoint := compose.CreateWebEndpoint(t, environment.WebEndpointCreate{
			UID:  device.UID,
			Host: "127.0.0.1",
			Port: first.tlsPort,
			TTL:  -1,
			TLS:  &environment.WebEndpointTLS{Enabled: true, Verify: true, Domain: deviceTLSDomain},
		})

		resp := visitWebEndpoint(t, compose, endpoint.FullAddress, "/")
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		assert.Contains(t, resp.Body, "failed to connect to the port on device")
	})

	t.Run("a subdomain that is not an address is not served", func(t *testing.T) {
		for _, path := range []string{"/api/info", "/internal/metrics"} {
			resp := visitWebEndpoint(t, compose, "console."+webEndpointDomain, path)

			assert.Equal(t, http.StatusNotFound, resp.StatusCode, "%s: %s", path, resp.Body)
		}
	})

	t.Run("a request to the instance's own domain still reaches the API", func(t *testing.T) {
		resp := visitWebEndpoint(t, compose, webEndpointDomain, "/api/info")

		assert.Equal(t, http.StatusOK, resp.StatusCode, resp.Body)
	})
}

func testWebEndpointLifetime(t *testing.T, compose *environment.DockerCompose, endless, elsewhere environment.WebEndpointCreate) {
	t.Helper()

	withTTL := func(ttl int) environment.WebEndpointCreate {
		req := endless
		req.TTL = ttl

		return req
	}

	t.Run("an endpoint never accessed expires its TTL after its creation", func(t *testing.T) {
		const ttl = 3600

		endpoint := listedWebEndpoint(t, compose, compose.CreateWebEndpoint(t, withTTL(ttl)).Address)

		assert.WithinDuration(t, endpoint.CreatedAt.Add(ttl*time.Second), endpoint.ExpiresIn, time.Millisecond)
	})

	t.Run("an accessed endpoint expires its TTL after its last access", func(t *testing.T) {
		const ttl = 3600

		endpoint := compose.CreateWebEndpoint(t, withTTL(ttl))
		created := listedWebEndpoint(t, compose, endpoint.Address)

		time.Sleep(2 * time.Second)

		before := clock.Now()
		requireDeviceAnswer(t, compose, endpoint.FullAddress, "/")
		after := clock.Now()

		accessed := listedWebEndpoint(t, compose, endpoint.Address)
		assert.True(t, accessed.ExpiresIn.After(created.ExpiresIn.Add(time.Second)), "the expiry %s did not move past %s", accessed.ExpiresIn, created.ExpiresIn)
		assert.WithinRange(t, accessed.ExpiresIn.Add(-ttl*time.Second), before.Add(-time.Second), after.Add(time.Second))
	})

	t.Run("an endpoint answers until its TTL runs out and is refused after it", func(t *testing.T) {
		endpoint := compose.CreateWebEndpoint(t, withTTL(shortTTL))

		requireDeviceAnswer(t, compose, endpoint.FullAddress, "/")
		requireRefusedOnceIdleFor(t, compose, endpoint.FullAddress, shortTTL)
	})

	t.Run("an endpoint accessed within its TTL keeps answering past it", func(t *testing.T) {
		const ttl = 5

		endpoint := compose.CreateWebEndpoint(t, withTTL(ttl))
		pastFirstExpiry := endpoint.CreatedAt.Add(ttl*time.Second + time.Second)

		for clock.Now().Before(pastFirstExpiry) {
			requireDeviceAnswer(t, compose, endpoint.FullAddress, "/")
			time.Sleep(time.Second)
		}

		requireDeviceAnswer(t, compose, endpoint.FullAddress, "/")
		requireRefusedOnceIdleFor(t, compose, endpoint.FullAddress, ttl)
	})

	t.Run("an endpoint with a TTL of -1 keeps answering after one with a TTL has expired", func(t *testing.T) {
		finite := compose.CreateWebEndpoint(t, withTTL(shortTTL))
		infinite := compose.CreateWebEndpoint(t, elsewhere)

		requireRefusedOnceIdleFor(t, compose, finite.FullAddress, shortTTL)

		requireDeviceAnswer(t, compose, infinite.FullAddress, "/")
		assert.True(t, listedWebEndpoint(t, compose, infinite.Address).ExpiresIn.IsZero())
	})

	t.Run("an expired endpoint can be deleted", func(t *testing.T) {
		endpoint := compose.CreateWebEndpoint(t, withTTL(shortTTL))
		requireRefusedOnceIdleFor(t, compose, endpoint.FullAddress, shortTTL)

		resp, err := compose.DeleteWebEndpoint(t.Context(), endpoint.Address)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

		for _, listed := range compose.ListWebEndpoints(t) {
			assert.NotEqual(t, endpoint.Address, listed.Address)
		}
	})
}

func testTunnelAPI(t *testing.T, compose *environment.DockerCompose, device *models.Device, service deviceService) {
	t.Helper()

	resp, err := compose.PostTunnel(t.Context(), device.UID, environment.TunnelCreate{Host: "127.0.0.1", Port: service.httpPort, TTL: -1})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

	tunnels := compose.ListTunnels(t, device.UID)
	require.Len(t, tunnels, 1)

	tunnel := tunnels[0]
	t.Cleanup(func() { _, _ = compose.DeleteTunnel(context.Background(), device.UID, tunnel.Address) })

	assert.Regexp(t, webEndpointAddress, tunnel.Address)
	assert.Equal(t, tunnel.Address+"."+webEndpointDomain, tunnel.FullAddress)
	assert.Equal(t, device.UID, tunnel.Device)
	assert.Equal(t, "127.0.0.1", tunnel.Host)
	assert.Equal(t, service.httpPort, tunnel.Port)
	assert.Equal(t, -1, tunnel.TTL)

	assert.Equal(t, service.name, requireDeviceAnswer(t, compose, tunnel.FullAddress, "/").Name)

	resp, err = compose.DeleteTunnel(t.Context(), device.UID, tunnel.Address)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

	assert.Empty(t, compose.ListTunnels(t, device.UID))
	requireEndpointRefused(t, compose, tunnel.FullAddress)
}

func testWebEndpointDeviceMerge(t *testing.T, ctx context.Context, compose *environment.DockerCompose, service deviceService) {
	t.Helper()

	same := []NewAgentContainerOption{NewAgentContainerWithIdentity("020000002202"), NewAgentContainerWithHostname("merging")}

	oldAgent, old := startAcceptedAgent(t, ctx, compose, same...)

	endpoint := compose.CreateWebEndpoint(t, environment.WebEndpointCreate{UID: old.UID, Host: "127.0.0.1", Port: service.httpPort, TTL: -1})
	requireDeviceAnswer(t, compose, endpoint.FullAddress, "/")

	require.NoError(t, oldAgent.Stop(ctx, nil))
	compose.AwaitDeviceOffline(t, old.UID)

	_, merged := startAcceptedAgent(t, ctx, compose, same...)
	require.NotEqual(t, old.UID, merged.UID)

	assert.Equal(t, merged.UID, listedWebEndpoint(t, compose, endpoint.Address).DeviceUID)
	assert.Equal(t, service.name, requireDeviceAnswer(t, compose, endpoint.FullAddress, "/").Name)
}

func startDeviceService(t *testing.T, compose *environment.DockerCompose, service deviceService) deviceService {
	t.Helper()

	gateway := compose.Service(environment.ServiceGateway).GetContainerID()

	buildArgs, err := testimage.BuildArgs("..")
	require.NoError(t, err)

	c, err := testcontainers.GenericContainer(t.Context(), testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			FromDockerfile: testcontainers.FromDockerfile{
				Repo:      "device-http",
				Tag:       "test",
				Context:   "cmd/device-http",
				KeepImage: true,
				BuildArgs: buildArgs,
			},
			Env: map[string]string{
				"NAME":         service.name,
				"HTTP_ADDRESS": "127.0.0.1:" + strconv.Itoa(service.httpPort),
				"TLS_ADDRESS":  "127.0.0.1:" + strconv.Itoa(service.tlsPort),
				"TLS_DOMAIN":   deviceTLSDomain,
			},
			Labels: run.Labels(),
			HostConfigModifier: func(hc *container.HostConfig) {
				hc.NetworkMode = container.NetworkMode("container:" + gateway)
			},
			WaitingFor: wait.ForLog("listening"),
		},
		Started: true,
		Logger:  log.New(io.Discard, "", log.LstdFlags),
	})
	require.NoError(t, err)

	t.Cleanup(func() { _ = c.Terminate(context.Background()) })

	return service
}

func gatewayAuthority(t *testing.T, compose *environment.DockerCompose) *x509.Certificate {
	t.Helper()

	var data []byte

	require.EventuallyWithT(t, func(tt *assert.CollectT) {
		reader, err := compose.Service(environment.ServiceGateway).CopyFileFromContainer(t.Context(), gatewayRootCertificate)
		if !assert.NoError(tt, err) {
			return
		}

		defer reader.Close() //nolint:errcheck // a failed read is retried; the close changes nothing

		data, err = io.ReadAll(reader)
		assert.NoError(tt, err)
	}, 30*time.Second, time.Second)

	block, _ := pem.Decode(data)
	require.NotNil(t, block, "the gateway's root certificate is not PEM")

	certificate, err := x509.ParseCertificate(block.Bytes)
	require.NoError(t, err)

	return certificate
}

func listedWebEndpoint(t *testing.T, compose *environment.DockerCompose, address string) environment.WebEndpoint {
	t.Helper()

	for _, endpoint := range compose.ListWebEndpoints(t) {
		if endpoint.Address == address {
			return endpoint
		}
	}

	require.Failf(t, "the web endpoint is not listed", "address %s", address)

	return environment.WebEndpoint{}
}

func requireCreateRefused(t *testing.T, compose *environment.DockerCompose, req environment.WebEndpointCreate, status int) {
	t.Helper()

	listed := len(compose.ListWebEndpoints(t))

	resp, err := compose.PostWebEndpoint(t.Context(), req, nil)
	require.NoError(t, err)
	assert.Equal(t, status, resp.StatusCode(), resp.String())
	assert.Len(t, compose.ListWebEndpoints(t), listed, "a refused create left an endpoint behind")
}

func visitWebEndpoint(t *testing.T, compose *environment.DockerCompose, host, uri string) *environment.WebEndpointResponse {
	t.Helper()

	resp, err := compose.VisitWebEndpoint(t.Context(), host, uri)
	require.NoError(t, err)

	return resp
}

func requireDeviceAnswer(t *testing.T, compose *environment.DockerCompose, host, uri string) deviceSeen {
	t.Helper()

	resp := visitWebEndpoint(t, compose, host, uri)
	require.Equal(t, http.StatusOK, resp.StatusCode, resp.Body)

	seen := deviceSeen{}
	require.NoError(t, json.Unmarshal([]byte(resp.Body), &seen), resp.Body)

	return seen
}

func requireEndpointRefused(t *testing.T, compose *environment.DockerCompose, host string) {
	t.Helper()

	resp := visitWebEndpoint(t, compose, host, "/")
	assert.Equal(t, http.StatusForbidden, resp.StatusCode, resp.Body)
	assert.Contains(t, resp.Body, webEndpointNotFound)
}

func requireRefusedOnceIdleFor(t *testing.T, compose *environment.DockerCompose, host string, ttl int) {
	t.Helper()

	time.Sleep(time.Duration(ttl)*time.Second + time.Second)

	requireEndpointRefused(t, compose, host)
}
