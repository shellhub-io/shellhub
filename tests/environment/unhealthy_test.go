package environment

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/testimage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/compose"
)

func TestUnhealthyServicesNamesEachFailedService(t *testing.T) {
	run := liveRun(t)

	root, err := testimage.Root()
	require.NoError(t, err)

	alpine, err := testimage.Version(root, "ALPINE_VERSION")
	require.NoError(t, err)

	for _, test := range []struct {
		name string
		file string
		want []string
	}{
		{
			name: "a service whose healthcheck never passes",
			file: "testdata/unhealthy.compose.yml",
			want: []string{"service sick is unhealthy", "probe-says-no"},
		},
		{
			name: "a service that exited with an error",
			file: "testdata/crashed.compose.yml",
			want: []string{"service crashed exited with code 3"},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			ctx := t.Context()

			project := "shellhub-e2e-unhealthy-" + newRunID()
			dc, err := newComposeStack(project, []string{test.file})
			require.NoError(t, err)

			t.Cleanup(func() { _ = dc.Down(context.Background(), compose.RemoveOrphans(true), compose.RemoveVolumes(true)) })

			envs := run.composeEnvs()
			envs["ALPINE_IMAGE"] = "alpine:" + alpine
			require.Error(t, dc.WithEnv(envs).Up(ctx, compose.Wait(true)))

			err = run.unhealthyServices(ctx, project)
			require.Error(t, err)

			for _, want := range test.want {
				assert.ErrorContains(t, err, want)
			}
		})
	}
}
