//go:build docker

package selfupdater

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	containertypes "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDockerContainerSplitImageVersion(t *testing.T) {
	cases := []struct {
		description string
		image       string
		repository  string
		version     string
	}{
		{
			description: "a Docker Hub image splits at its tag",
			image:       "shellhubio/agent:v0.21.0",
			repository:  "shellhubio/agent",
			version:     "v0.21.0",
		},
		{
			description: "a registry with a port keeps the port in the repository",
			image:       "registry.example.com:5000/shellhubio/agent:v0.21.0",
			repository:  "registry.example.com:5000/shellhubio/agent",
			version:     "v0.21.0",
		},
		{
			description: "a registry with a port and no tag has no version",
			image:       "localhost:5000/agent",
			repository:  "localhost:5000/agent",
			version:     "",
		},
		{
			description: "an image without a tag has no version",
			image:       "shellhubio/agent",
			repository:  "shellhubio/agent",
			version:     "",
		},
		{
			description: "a digest is not taken for the version",
			image:       "registry.example.com:5000/shellhubio/agent:v0.21.0@sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
			repository:  "registry.example.com:5000/shellhubio/agent",
			version:     "v0.21.0",
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			container := &dockerContainer{info: &containertypes.InspectResponse{Config: &containertypes.Config{Image: tc.image}}}

			repository, version := container.splitImageVersion()
			assert.Equal(t, tc.repository, repository)
			assert.Equal(t, tc.version, version)
		})
	}
}

func TestAllowInsecureServerForExistingHTTP(t *testing.T) {
	cases := []struct {
		description string
		env         []string
		expected    []string
	}{
		{
			description: "an http address off loopback gets the insecure server allowed",
			env:         []string{"SHELLHUB_SERVER_ADDRESS=http://192.0.2.10:8080", "SHELLHUB_TENANT_ID=tenant"},
			expected:    []string{"SHELLHUB_SERVER_ADDRESS=http://192.0.2.10:8080", "SHELLHUB_TENANT_ID=tenant", "SHELLHUB_ALLOW_INSECURE_SERVER=true"},
		},
		{
			description: "an http address under the unprefixed key gets the insecure server allowed",
			env:         []string{"SERVER_ADDRESS=http://192.0.2.10"},
			expected:    []string{"SERVER_ADDRESS=http://192.0.2.10", "SHELLHUB_ALLOW_INSECURE_SERVER=true"},
		},
		{
			description: "the prefixed address wins over the unprefixed one",
			env:         []string{"SHELLHUB_SERVER_ADDRESS=https://shellhub.example.com", "SERVER_ADDRESS=http://192.0.2.10"},
			expected:    []string{"SHELLHUB_SERVER_ADDRESS=https://shellhub.example.com", "SERVER_ADDRESS=http://192.0.2.10"},
		},
		{
			description: "an empty choice on the insecure server is replaced",
			env:         []string{"SHELLHUB_SERVER_ADDRESS=http://192.0.2.10", "SHELLHUB_ALLOW_INSECURE_SERVER="},
			expected:    []string{"SHELLHUB_SERVER_ADDRESS=http://192.0.2.10", "SHELLHUB_ALLOW_INSECURE_SERVER=true"},
		},
		{
			description: "a value the agent does not read as true is replaced",
			env:         []string{"SHELLHUB_SERVER_ADDRESS=http://192.0.2.10", "SHELLHUB_ALLOW_INSECURE_SERVER=1"},
			expected:    []string{"SHELLHUB_SERVER_ADDRESS=http://192.0.2.10", "SHELLHUB_ALLOW_INSECURE_SERVER=true"},
		},
		{
			description: "an explicit true in any case is kept",
			env:         []string{"SHELLHUB_SERVER_ADDRESS=http://192.0.2.10", "SHELLHUB_ALLOW_INSECURE_SERVER=TRUE"},
			expected:    []string{"SHELLHUB_SERVER_ADDRESS=http://192.0.2.10", "SHELLHUB_ALLOW_INSECURE_SERVER=TRUE"},
		},
		{
			description: "an invalid address is left alone",
			env:         []string{"SHELLHUB_SERVER_ADDRESS=not a url"},
			expected:    []string{"SHELLHUB_SERVER_ADDRESS=not a url"},
		},
		{
			description: "an https address is left alone",
			env:         []string{"SHELLHUB_SERVER_ADDRESS=https://shellhub.example.com"},
			expected:    []string{"SHELLHUB_SERVER_ADDRESS=https://shellhub.example.com"},
		},
		{
			description: "an http loopback address is left alone",
			env:         []string{"SHELLHUB_SERVER_ADDRESS=http://localhost"},
			expected:    []string{"SHELLHUB_SERVER_ADDRESS=http://localhost"},
		},
		{
			description: "an explicit choice on the insecure server is kept",
			env:         []string{"SHELLHUB_SERVER_ADDRESS=http://192.0.2.10", "SHELLHUB_ALLOW_INSECURE_SERVER=false"},
			expected:    []string{"SHELLHUB_SERVER_ADDRESS=http://192.0.2.10", "SHELLHUB_ALLOW_INSECURE_SERVER=false"},
		},
		{
			description: "no server address is left alone",
			env:         []string{"SHELLHUB_TENANT_ID=tenant"},
			expected:    []string{"SHELLHUB_TENANT_ID=tenant"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.Equal(t, tc.expected, allowInsecureServerForExistingHTTP(tc.env))
		})
	}
}

func TestCompleteUpdateAllowsInsecureServerOnTheReplacement(t *testing.T) {
	const parentID = "parent"

	t.Setenv("PARENT_CONTAINER", parentID)

	var created containertypes.Config

	engine := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/json"):
			image := "shellhubio/agent:v0.31.1"
			if strings.Contains(r.URL.Path, "/containers/"+parentID+"/") {
				image = "shellhubio/agent:v0.30.3"
			}

			_ = json.NewEncoder(w).Encode(containertypes.InspectResponse{
				ContainerJSONBase: &containertypes.ContainerJSONBase{ID: parentID, Name: "/shellhub", HostConfig: &containertypes.HostConfig{}},
				Config: &containertypes.Config{
					Image: image,
					Env:   []string{"SHELLHUB_SERVER_ADDRESS=http://192.0.2.10:8080"},
				},
				NetworkSettings: &containertypes.NetworkSettings{Networks: map[string]*network.EndpointSettings{}},
			})
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/containers/create"):
			if err := json.NewDecoder(r.Body).Decode(&created); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)

				return
			}

			_ = json.NewEncoder(w).Encode(containertypes.CreateResponse{ID: "replacement"})
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/images/create"):
			_, _ = w.Write([]byte("{}"))
		default:
			w.WriteHeader(http.StatusNoContent)
		}
	}))
	t.Cleanup(engine.Close)

	api, err := client.NewClientWithOpts(client.WithHost("tcp://"+strings.TrimPrefix(engine.URL, "http://")), client.WithVersion("1.47"))
	require.NoError(t, err)

	require.NoError(t, (&dockerUpdater{api: api, containerID: func() (string, error) { return "replacement", nil }}).CompleteUpdate())
	assert.Contains(t, created.Env, "SHELLHUB_ALLOW_INSECURE_SERVER=true")
}
