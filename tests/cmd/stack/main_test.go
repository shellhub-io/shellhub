package main

import (
	"bytes"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultStackName(t *testing.T) {
	cases := []struct {
		name  string
		roots []string
		same  bool
	}{
		{
			name:  "the same checkout gets the same name",
			roots: []string{"/home/dev/src/shellhub-io/shellhub", "/home/dev/src/shellhub-io/shellhub"},
			same:  true,
		},
		{
			name:  "two worktrees get different names",
			roots: []string{"/home/dev/.t3/worktrees/a/shellhub", "/home/dev/.t3/worktrees/b/shellhub"},
			same:  false,
		},
		{
			name:  "a path that would not be a valid name still gives one",
			roots: []string{"/Users/Dev User/Shell Hub", "/"},
			same:  false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			first := defaultStackName(tc.roots[0])
			second := defaultStackName(tc.roots[1])

			assert.Regexp(t, validName, first)
			assert.Regexp(t, validName, second)
			assert.Equal(t, tc.same, first == second)
		})
	}
}

func TestPrintExportsSurvivesEval(t *testing.T) {
	state := &stateFile{Name: "shellhub-e2e-a", Edition: "enterprise", AgentImage: "agent:a b"}

	var out bytes.Buffer
	require.NoError(t, printExports(&out, state, []byte(`it's "$HOME" * ;`)))

	script := out.String() + `printf '%s|%s' "$E2E_AGENT_IMAGE" "$E2E_EXPIRED_LICENSE"`
	sh := exec.CommandContext(t.Context(), "sh")
	sh.Stdin = strings.NewReader(script)
	got, err := sh.Output()
	require.NoError(t, err)
	assert.Equal(t, `agent:a b|it's "$HOME" * ;`, string(got))
}
