package main

import (
	"context"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	samlIdentityProvider = "http://idp.e2e.test"
	samlRelayTokenTTL    = 5 * time.Minute
)

// TestEnterpriseSAMLReauth covers the relay token behind a SAML user's re-authentication for the
// web terminal: the server issues it in the URL that sends the user to the identity provider, keeps
// it for five minutes, and refuses it once it expires. The provider takes the HTTP-POST binding, so
// the server checks the token when it builds the form posted to the provider, and no provider has
// to answer.
func TestEnterpriseSAMLReauth(t *testing.T) {
	t.Parallel()

	compose := newEnterpriseEnvironment(t, context.Background(), environment.New(t, run))
	enableSAMLPostBinding(t, compose)

	t.Run("a relay token lasts five minutes", func(t *testing.T) {
		token := issueSAMLRelayToken(t, compose)

		left := compose.CacheEntryTTL(t, samlRelayTokenCacheKey(token))
		assert.Greater(t, left, samlRelayTokenTTL-time.Minute)
		assert.LessOrEqual(t, left, samlRelayTokenTTL)

		assert.Equal(t, http.StatusOK, requestSAMLReauth(t, compose, token).StatusCode())
	})

	t.Run("an expired relay token is refused", func(t *testing.T) {
		token := issueSAMLRelayToken(t, compose)

		compose.ExpireCacheEntryIn(t, samlRelayTokenCacheKey(token), 0)

		assert.Equal(t, http.StatusForbidden, requestSAMLReauth(t, compose, token).StatusCode())
	})
}

func enableSAMLPostBinding(t *testing.T, compose *environment.DockerCompose) {
	t.Helper()

	resp, err := compose.R(t.Context()).
		SetBody(map[string]any{
			"enable": true,
			"idp": map[string]any{
				"entity_id":   samlIdentityProvider,
				"certificate": samlIdentityProviderCertificate(t),
				"binding":     map[string]string{"post": samlIdentityProvider + "/sso/post"},
			},
			"sp": map[string]bool{"sign_requests": false},
		}).
		Put("/admin/api/authentication/saml")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())
}

func samlIdentityProviderCertificate(t *testing.T) string {
	t.Helper()

	certificate, err := environment.SelfSignedCertificate("idp.e2e.test")
	require.NoError(t, err)

	return string(certificate.CertificatePEM)
}

func samlRelayTokenCacheKey(token string) string {
	return "saml-stepup/" + token
}

func issueSAMLRelayToken(t *testing.T, compose *environment.DockerCompose) string {
	t.Helper()

	reauth := struct {
		URL string `json:"url"`
	}{}

	resp, err := compose.R(t.Context()).
		SetQueryParams(map[string]string{"fingerprint": "SHA256:e2e", "approval_code": "E2E"}).
		SetResult(&reauth).
		Get("/api/user/saml/reauth")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

	parsed, err := url.Parse(reauth.URL)
	require.NoError(t, err)

	token := parsed.Query().Get("relay_state")
	require.NotEmpty(t, token, "the re-authentication URL %s carries no relay token", reauth.URL)

	return token
}

func requestSAMLReauth(t *testing.T, compose *environment.DockerCompose, token string) *resty.Response {
	t.Helper()

	resp, err := compose.Anonymous(t.Context()).
		SetQueryParam("relay_state", token).
		Get("/api/user/saml/request")
	require.NoError(t, err)

	return resp
}
