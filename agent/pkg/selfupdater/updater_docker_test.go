//go:build docker

package selfupdater

import (
	"testing"

	containertypes "github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/assert"
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
