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

func TestAgentImageKey(t *testing.T) {
	repoFiles := map[string]string{
		"versions.env":                      "GOLANG_VERSION=1\n",
		".dockerignore":                     ".git\n",
		"go.mod":                            "module github.com/shellhub-io/shellhub\n",
		"go.sum":                            "",
		"pkg/a.go":                          "package pkg\n",
		"agent/Dockerfile.test":             "FROM scratch\n",
		"agent/main.go":                     "package main\n",
		"tests/cmd/signal-reporter/main.go": "package main\n",
		"server/main.go":                    "package main\n",
	}

	keyOf := func(t *testing.T, repo string, build AgentBuild) string {
		t.Helper()

		got, err := agentImageKey(t.Context(), repo, build)
		require.NoError(t, err)

		return got
	}

	cases := []struct {
		name    string
		repo    map[string]string
		build   AgentBuild
		changed bool
	}{
		{name: "a file under the agent", repo: map[string]string{"agent/main.go": "package main // changed\n"}, changed: true},
		{name: "the shared packages", repo: map[string]string{"pkg/b.go": "package pkg\n"}, changed: true},
		{name: "the root module", repo: map[string]string{"go.sum": "changed\n"}, changed: true},
		{name: "the signal reporter", repo: map[string]string{"tests/cmd/signal-reporter/main.go": "package main // changed\n"}, changed: true},
		{name: "the Dockerfile", repo: map[string]string{"agent/Dockerfile.test": "FROM alpine\n"}, changed: true},
		{name: "versions.env", repo: map[string]string{"versions.env": "GOLANG_VERSION=2\n"}, changed: true},
		{name: "the version", build: AgentBuild{Version: "0.20.0"}, changed: true},
		{name: "the build tags", build: AgentBuild{Tags: []string{"docker"}}, changed: true},
		{name: "a file outside the agent's paths", repo: map[string]string{"server/main.go": "package main // changed\n"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := gitFixture(t, repoFiles)

			before := keyOf(t, repo, AgentBuild{})

			if tc.repo != nil {
				commitFiles(t, repo, tc.repo)
			}

			after := keyOf(t, repo, tc.build)

			if tc.changed {
				assert.NotEqual(t, before, after)
			} else {
				assert.Equal(t, before, after)
			}
		})
	}
}
