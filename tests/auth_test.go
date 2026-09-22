package main

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLoginWithoutANamespace logs in a user who holds no membership. The stack runs the API under
// SHELLHUB_OPENAPI_VALIDATION=strict, so the 200 carries the real assertion: a body the schema
// rejects is replaced by a 500. Decoding into a map rather than [models.UserAuthResponse] keeps a
// dropped key from reading as a null one.
func TestLoginWithoutANamespace(t *testing.T) {
	ctx := context.Background()

	compose := environment.New(t).Up(ctx)
	t.Cleanup(compose.Down)

	compose.NewUser(t, "nobody", "nobody@ossystems.com.br", ShellHubPassword)

	body := map[string]any{}

	require.EventuallyWithT(t, func(tt *assert.CollectT) {
		resp, err := compose.R(ctx).
			SetBody(map[string]string{
				"username": "nobody",
				"password": ShellHubPassword,
			}).
			Post("/api/login")
		assert.NoError(tt, err)
		assert.Equal(tt, http.StatusOK, resp.StatusCode())
		assert.NoError(tt, json.Unmarshal(resp.Body(), &body))
	}, 30*time.Second, 1*time.Second)

	for _, key := range []string{"tenant", "role"} {
		require.Contains(t, body, key)
		assert.Nil(t, body[key])
	}
}
