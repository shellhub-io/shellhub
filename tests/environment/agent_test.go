package environment

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildAgentImage(t *testing.T) {
	run := liveRun(t)

	t.Run("builds with BuildKit from the repository context", func(t *testing.T) {
		var out bytes.Buffer

		image, err := BuildAgentImage(t.Context(), run, AgentBuild{
			Repository: "shellhub-e2e/agent-build-test",
			Context:    "../..",
			Dockerfile: "tests/environment/testdata/cache-mount.Dockerfile",
			Output:     &out,
		})
		require.NoError(t, err)
		assert.Equal(t, "shellhub-e2e/agent-build-test:"+run.ID(), image)

		_, err = run.client.ImageInspect(t.Context(), image)
		require.NoError(t, err)

		assert.Regexp(t, `(?m)^#\d+ \[stage-0 2/2\] RUN --mount=type=cache,id=shellhub-e2e-agent-test`, out.String())
	})

	t.Run("returns the daemon's message when a step fails", func(t *testing.T) {
		_, err := BuildAgentImage(t.Context(), run, AgentBuild{
			Repository: "shellhub-e2e/agent-build-test",
			Context:    "../..",
			Dockerfile: "tests/environment/testdata/failing.Dockerfile",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "exit code: 3")
	})
}
