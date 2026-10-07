package main

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const sessionRetentionCron = "0 1 * * *"

func TestSessionRetention(t *testing.T) {
	ctx := context.Background()

	compose := newConfiguredSSHEnvironment(t, ctx,
		environment.New(t, run).WithEnv("SHELLHUB_SESSION_RETENTION_DAYS", "1").WithCronTrigger(),
		"legacy")
	compose.AwaitServerLog(t, "session retention enabled")

	_, device := startAcceptedAgent(t, ctx, compose)
	signer := registerDeviceKey(t, ctx, compose)

	expired := finishSession(t, ctx, compose, device, signer, "echo expired")
	recent := finishSession(t, ctx, compose, device, signer, "echo recent")

	_, live := openSession(t, ctx, compose, device, signer, "")
	requireSessionActive(t, ctx, compose, live.UID, true)

	compose.AgeSession(t, expired, 48*time.Hour)
	compose.AgeSession(t, live.UID, 48*time.Hour)

	require.Positive(t, compose.SessionEventCount(t, expired), "the expired session recorded events for the cascade to take")

	compose.RunCron(t, sessionRetentionCron)

	awaitSessionDeleted(t, ctx, compose, expired)

	compose.AwaitServerLog(t, "pruned sessions past the retention window")

	assert.Zero(t, compose.SessionEventCount(t, expired), "the expired session's events go with it")

	t.Run("a session inside the window is kept", func(t *testing.T) {
		assert.Equal(t, recent, getSession(t, ctx, compose, recent).UID)
		assert.Positive(t, compose.SessionEventCount(t, recent))
	})

	t.Run("a session past the window that is still open is kept", func(t *testing.T) {
		session := getSession(t, ctx, compose, live.UID)
		assert.True(t, session.Active)
		assert.Positive(t, compose.SessionEventCount(t, live.UID))
	})
}

func awaitSessionDeleted(t *testing.T, ctx context.Context, compose *environment.DockerCompose, uid string) {
	t.Helper()

	require.EventuallyWithT(t, func(tt *assert.CollectT) {
		resp, err := compose.R(ctx).Get("/api/sessions/" + uid)
		assert.NoError(tt, err)
		assert.Equal(tt, http.StatusNotFound, resp.StatusCode(), "session %s is still there", uid)
	}, 30*time.Second, 1*time.Second)
}
