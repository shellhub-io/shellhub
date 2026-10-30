package main

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOnboardingSurvey(t *testing.T) {
	t.Parallel()

	compose := newSSHEnvironment(t, t.Context(), legacyNamespace)

	t.Run("a new account starts pending, on sign-in and on the user info", func(t *testing.T) {
		compose.NewUser(t, "surveyfresh", "surveyfresh@ossystems.com.br", ShellHubPassword)

		login, err := compose.R(t.Context()).
			SetBody(map[string]string{"username": "surveyfresh", "password": ShellHubPassword}).
			Post("/api/login")
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, login.StatusCode(), login.String())
		assert.Equal(t, "pending", onboardingSurveyOf(t, login))

		token := compose.AuthUser(t, "surveyfresh", ShellHubPassword).Token
		assert.Equal(t, "pending", currentOnboardingSurvey(t, compose, token))
	})

	for _, outcome := range []string{"answered", "skipped"} {
		t.Run("recording "+outcome+" is read back", func(t *testing.T) {
			username := "survey" + outcome
			compose.NewUser(t, username, username+"@ossystems.com.br", ShellHubPassword)
			token := compose.AuthUser(t, username, ShellHubPassword).Token

			resp, err := recordOnboardingSurvey(asBearer(t, compose, token), outcome)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

			assert.Equal(t, outcome, currentOnboardingSurvey(t, compose, token))
			assert.Equal(t, outcome, currentOnboardingSurvey(t, compose, compose.AuthUser(t, username, ShellHubPassword).Token))
		})
	}

	t.Run("an outcome other than answered or skipped is refused", func(t *testing.T) {
		compose.NewUser(t, "surveyinvalid", "surveyinvalid@ossystems.com.br", ShellHubPassword)
		token := compose.AuthUser(t, "surveyinvalid", ShellHubPassword).Token

		for _, outcome := range []string{"pending", "maybe", ""} {
			resp, err := recordOnboardingSurvey(asBearer(t, compose, token), outcome)
			require.NoError(t, err)
			assert.Equal(t, http.StatusBadRequest, resp.StatusCode(), "outcome %q: %s", outcome, resp.String())
		}

		assert.Equal(t, "pending", currentOnboardingSurvey(t, compose, token))
	})

	t.Run("an API key cannot record an outcome", func(t *testing.T) {
		key := compose.CreateAPIKey(t, &requests.CreateAPIKey{Name: "survey", ExpiresAt: -1})

		resp, err := recordOnboardingSurvey(compose.Anonymous(t.Context()).SetHeader("X-API-Key", key.Key), "skipped")
		require.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, resp.StatusCode(), resp.String())
	})
}

func recordOnboardingSurvey(req *resty.Request, outcome string) (*resty.Response, error) {
	return req.SetBody(map[string]string{"outcome": outcome}).Put("/api/users/onboarding-survey")
}

func currentOnboardingSurvey(t *testing.T, compose *environment.DockerCompose, token string) string {
	t.Helper()

	resp, err := asBearer(t, compose, token).Get("/api/auth/user")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

	return onboardingSurveyOf(t, resp)
}

func onboardingSurveyOf(t *testing.T, resp *resty.Response) string {
	t.Helper()

	body := map[string]any{}
	require.NoError(t, json.Unmarshal(resp.Body(), &body))
	require.Contains(t, body, "onboarding_survey")

	outcome, ok := body["onboarding_survey"].(string)
	require.True(t, ok, "onboarding_survey is %T", body["onboarding_survey"])

	return outcome
}
