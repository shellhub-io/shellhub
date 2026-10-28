package main

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCloudDisposableEmail runs a cloud instance whose server asks the stack's stub DNS server,
// which answers an MX record for every domain but nomx.test, whether an email domain can receive
// mail, so no case reaches the network.
func TestCloudDisposableEmail(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("checked", func(t *testing.T) {
		t.Parallel()

		compose := environment.New(t, run).
			WithEdition(environment.EditionCloud).
			WithEnv("SHELLHUB_DISPOSABLE_EMAIL_CHECK", "true").
			Up(ctx)
		t.Cleanup(compose.Down)

		cases := []struct {
			description string
			email       string
			reason      string
		}{
			{"refuses a listed domain", "listed@mailinator.com", "disposable"},
			{"refuses a subdomain of a listed domain", "subdomain@x.mailinator.com", "disposable"},
			{"refuses a domain that cannot receive mail", "nomx@nomx.test", "unreachable"},
		}

		for _, tc := range cases {
			t.Run("sign-up "+tc.description, func(t *testing.T) {
				resp := signUpWithEmail(t, compose, uuid.Generate()[:8], tc.email)
				require.Equal(t, http.StatusBadRequest, resp.StatusCode(), resp.String())
				assert.Equal(t, map[string]string{"email": tc.reason}, refusedFields(t, resp))
			})
		}

		t.Run("sign-up accepts a plus alias on a domain that receives mail", func(t *testing.T) {
			resp := signUpWithEmail(t, compose, "plus-alias", "me+shellhub"+stripeTestEmailDomain)
			assert.Equal(t, http.StatusNoContent, resp.StatusCode(), resp.String())
		})

		username := "email-change"
		compose.NewUser(t, username, username+stripeTestEmailDomain, ShellHubPassword)
		compose.NewNamespace(t, username, username, uuid.Generate())
		token := compose.AuthUser(t, username, ShellHubPassword).Token

		for _, field := range []string{"email", "recovery_email"} {
			t.Run("an update refuses a listed "+field, func(t *testing.T) {
				resp, err := asBearer(t, compose, token).
					SetBody(map[string]string{field: field + "@mailinator.com"}).
					Patch("/api/users")
				require.NoError(t, err)
				require.Equal(t, http.StatusBadRequest, resp.StatusCode(), resp.String())
				assert.Equal(t, map[string]string{field: "disposable"}, refusedFields(t, resp))
			})
		}
	})

	t.Run("turned off", func(t *testing.T) {
		t.Parallel()

		compose := environment.New(t, run).
			WithEdition(environment.EditionCloud).
			WithEnv("SHELLHUB_DISPOSABLE_EMAIL_CHECK", "false").
			Up(ctx)
		t.Cleanup(compose.Down)

		resp := signUpWithEmail(t, compose, "listed-off", "listed-off@mailinator.com")
		assert.Equal(t, http.StatusNoContent, resp.StatusCode(), resp.String())
	})
}

func signUpWithEmail(t *testing.T, compose *environment.DockerCompose, username, email string) *resty.Response {
	t.Helper()

	resp, err := compose.Anonymous(t.Context()).
		SetBody(map[string]string{
			"name":     username,
			"username": username,
			"email":    email,
			"password": ShellHubPassword,
		}).
		Post("/api/register")
	require.NoError(t, err)

	return resp
}

func refusedFields(t *testing.T, resp *resty.Response) map[string]string {
	t.Helper()

	var body struct {
		Fields map[string]string `json:"fields"`
	}
	require.NoError(t, json.Unmarshal(resp.Body(), &body), resp.String())

	return body.Fields
}
