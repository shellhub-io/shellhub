package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func sessionFilter(t *testing.T, name, operator string, value any) string {
	t.Helper()

	encoded, err := json.Marshal([]query.Filter{{
		Type:   query.FilterTypeProperty,
		Params: &query.FilterProperty{Name: name, Operator: operator, Value: value},
	}})
	require.NoError(t, err)

	return base64.RawURLEncoding.EncodeToString(encoded)
}

func requireSessionsListed(t *testing.T, ctx context.Context, compose *environment.DockerCompose, filter string, want ...string) {
	t.Helper()

	sessions := []models.Session{}

	resp, err := compose.R(ctx).
		SetQueryParams(map[string]string{"filter": filter, "per_page": "100"}).
		SetResult(&sessions).
		Get("/api/sessions")
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode(), resp.String())

	uids := make([]string, 0, len(sessions))
	for _, session := range sessions {
		uids = append(uids, session.UID)
	}

	assert.ElementsMatch(t, want, uids)
}

func TestSessionList(t *testing.T) {
	ctx := context.Background()

	compose := newSSHEnvironment(t, ctx, "legacy")
	_, first := startAcceptedAgent(t, ctx, compose)
	_, second := startAcceptedAgent(t, ctx, compose)
	signer := registerDeviceKey(t, ctx, compose)

	oldest := finishSession(t, ctx, compose, first, signer, "true")
	onSecond := finishSession(t, ctx, compose, second, signer, "true")

	_, opened := openSession(t, ctx, compose, first, signer, "")
	live := opened.UID
	requireSessionActive(t, ctx, compose, live, true)

	t.Run("by device_uid", func(t *testing.T) {
		requireSessionsListed(t, ctx, compose, sessionFilter(t, "device_uid", "eq", first.UID), oldest, live)
		requireSessionsListed(t, ctx, compose, sessionFilter(t, "device_uid", "eq", second.UID), onSecond)
		requireSessionsListed(t, ctx, compose, sessionFilter(t, "device_uid", "ne", first.UID), onSecond)
	})

	t.Run("by active status", func(t *testing.T) {
		requireSessionsListed(t, ctx, compose, sessionFilter(t, "active", "bool", true), live)
		requireSessionsListed(t, ctx, compose, sessionFilter(t, "active", "bool", false), oldest, onSecond)
	})

	t.Run("by closed status", func(t *testing.T) {
		requireSessionsListed(t, ctx, compose, sessionFilter(t, "closed", "bool", true), oldest, onSecond)
		requireSessionsListed(t, ctx, compose, sessionFilter(t, "closed", "bool", false), live)
	})

	t.Run("newest started first", func(t *testing.T) {
		sessions := []models.Session{}

		resp, err := compose.R(ctx).SetResult(&sessions).Get("/api/sessions")
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode(), resp.String())

		uids := make([]string, 0, len(sessions))
		for i, session := range sessions {
			uids = append(uids, session.UID)

			if i > 0 {
				assert.False(t, session.StartedAt.After(sessions[i-1].StartedAt), "the list is newest first")
			}
		}

		assert.Equal(t, []string{live, onSecond, oldest}, uids)
	})
}
