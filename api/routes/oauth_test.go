package routes

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	servicemock "github.com/shellhub-io/shellhub/api/services/mocks"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/jwttoken"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// oauthTestKeys generates an RSA keypair for tests that need to sign JWTs.
func oauthTestKeys(t *testing.T) (*rsa.PrivateKey, *rsa.PublicKey) {
	t.Helper()
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	return priv, &priv.PublicKey
}

// oauthBearer encodes the given claims into a signed bearer token suitable for
// the Authorization header.
func oauthBearer(t *testing.T, priv *rsa.PrivateKey, claims authorizer.UserClaims) string {
	t.Helper()
	token, err := jwttoken.EncodeUserClaims(claims, priv)
	require.NoError(t, err)

	return "Bearer " + token
}

func TestOAuthMetadata(t *testing.T) {
	svcMock := new(servicemock.Service)

	req := httptest.NewRequest(http.MethodGet, "/.well-known/oauth-authorization-server", nil)
	rec := httptest.NewRecorder()
	e := NewRouter(svcMock)
	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var body map[string]any
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
	require.Contains(t, body, "issuer")
	require.Contains(t, body, "authorization_endpoint")
	require.Contains(t, body, "token_endpoint")
	require.Contains(t, body, "code_challenge_methods_supported")
}

func TestOAuthAuthorize(t *testing.T) {
	type Expected struct {
		status   int
		location string
	}

	svcMock := new(servicemock.Service)

	cases := []struct {
		description   string
		query         url.Values
		requiredMocks func()
		expected      Expected
	}{
		{
			description: "fails when code_challenge_method is not S256",
			query: url.Values{
				"client_id":             {"client-id"},
				"redirect_uri":          {"http://localhost/callback"},
				"code_challenge":        {"abc123"},
				"code_challenge_method": {"plain"},
				"state":                 {"xyz"},
			},
			requiredMocks: func() {},
			expected:      Expected{status: http.StatusBadRequest},
		},
		{
			description: "fails when client_id is unknown",
			query: url.Values{
				"client_id":             {"unknown-client"},
				"redirect_uri":          {"http://localhost/callback"},
				"code_challenge":        {"abc123"},
				"code_challenge_method": {"S256"},
				"state":                 {"xyz"},
			},
			requiredMocks: func() {
				svcMock.
					On("OAuthGetClient", mock.Anything, "unknown-client").
					Return(nil, errors.New("not found")).
					Once()
			},
			expected: Expected{status: http.StatusBadRequest},
		},
		{
			description: "fails when redirect_uri not in allowed list",
			query: url.Values{
				"client_id":             {"client-id"},
				"redirect_uri":          {"http://evil.example/callback"},
				"code_challenge":        {"abc123"},
				"code_challenge_method": {"S256"},
				"state":                 {"xyz"},
			},
			requiredMocks: func() {
				svcMock.
					On("OAuthGetClient", mock.Anything, "client-id").
					Return(&models.OAuthClient{
						ClientID:     "client-id",
						RedirectURIs: []string{"http://localhost/callback"},
					}, nil).
					Once()
			},
			expected: Expected{status: http.StatusBadRequest},
		},
		{
			description: "succeeds and redirects to login",
			query: url.Values{
				"client_id":             {"client-id"},
				"redirect_uri":          {"http://localhost/callback"},
				"code_challenge":        {"abc123"},
				"code_challenge_method": {"S256"},
				"state":                 {"xyz"},
			},
			requiredMocks: func() {
				svcMock.
					On("OAuthGetClient", mock.Anything, "client-id").
					Return(&models.OAuthClient{
						ClientID:     "client-id",
						RedirectURIs: []string{"http://localhost/callback"},
					}, nil).
					Once()
			},
			expected: Expected{
				status:   http.StatusFound,
				location: "/login",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			req := httptest.NewRequest(http.MethodGet, "/api/oauth/authorize?"+tc.query.Encode(), nil)
			rec := httptest.NewRecorder()
			e := NewRouter(svcMock)
			e.ServeHTTP(rec, req)

			require.Equal(t, tc.expected.status, rec.Code)
			if tc.expected.location != "" {
				require.Contains(t, rec.Header().Get("Location"), tc.expected.location)
			}
		})
	}
}

func TestOAuthToken(t *testing.T) {
	type Expected struct {
		status int
	}

	svcMock := new(servicemock.Service)

	cases := []struct {
		description   string
		form          url.Values
		requiredMocks func()
		expected      Expected
	}{
		{
			description: "fails when grant_type is not authorization_code",
			form: url.Values{
				"grant_type": {"client_credentials"},
			},
			requiredMocks: func() {},
			expected:      Expected{status: http.StatusBadRequest},
		},
		{
			description: "fails when code is invalid",
			form: url.Values{
				"grant_type":    {"authorization_code"},
				"code":          {"badcode"},
				"client_id":     {"client-id"},
				"redirect_uri":  {"http://localhost/callback"},
				"code_verifier": {"verifier"},
			},
			requiredMocks: func() {
				svcMock.
					On("OAuthExchangeCode", mock.Anything, "badcode", "client-id", "http://localhost/callback", "verifier").
					Return(nil, errors.New("invalid code")).
					Once()
			},
			expected: Expected{status: http.StatusUnauthorized},
		},
		{
			description: "succeeds",
			form: url.Values{
				"grant_type":    {"authorization_code"},
				"code":          {"validcode"},
				"client_id":     {"client-id"},
				"redirect_uri":  {"http://localhost/callback"},
				"code_verifier": {"verifier"},
			},
			requiredMocks: func() {
				svcMock.
					On("OAuthExchangeCode", mock.Anything, "validcode", "client-id", "http://localhost/callback", "verifier").
					Return(&models.UserAuthResponse{Token: "jwt-token"}, nil).
					Once()
			},
			expected: Expected{status: http.StatusOK},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			req := httptest.NewRequest(http.MethodPost, "/api/oauth/token", strings.NewReader(tc.form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			rec := httptest.NewRecorder()
			e := NewRouter(svcMock)
			e.ServeHTTP(rec, req)

			require.Equal(t, tc.expected.status, rec.Code)
		})
	}
}

func TestOAuthCallback(t *testing.T) {
	priv, pub := oauthTestKeys(t)
	tenant := "00000000-0000-4000-0000-000000000000"
	bearer := oauthBearer(t, priv, authorizer.UserClaims{ID: "user-1", TenantID: tenant})

	t.Run("fails when bearer token is missing", func(t *testing.T) {
		svcMock := new(servicemock.Service)
		svcMock.On("PublicKey").Return(pub).Maybe()

		body := map[string]any{
			"client_id":      "client-id",
			"redirect_uri":   "http://localhost/cb",
			"code_challenge": "challenge",
			"state":          "state",
		}
		data, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/oauth/authorize/callback", strings.NewReader(string(data)))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		NewRouter(svcMock).ServeHTTP(rec, req)
		require.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("fails when redirect_uri is not registered for client", func(t *testing.T) {
		svcMock := new(servicemock.Service)
		svcMock.On("PublicKey").Return(pub).Maybe()
		svcMock.
			On("OAuthGetClient", mock.Anything, "client-id").
			Return(&models.OAuthClient{ClientID: "client-id", RedirectURIs: []string{"https://app/cb"}}, nil).
			Once()

		body := map[string]any{
			"client_id":      "client-id",
			"redirect_uri":   "https://attacker/cb",
			"code_challenge": "challenge",
			"state":          "state",
		}
		data, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/oauth/authorize/callback", strings.NewReader(string(data)))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", bearer)
		rec := httptest.NewRecorder()
		NewRouter(svcMock).ServeHTTP(rec, req)
		require.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("issues a code with claims tenant when authenticated and redirect matches", func(t *testing.T) {
		svcMock := new(servicemock.Service)
		svcMock.On("PublicKey").Return(pub).Maybe()
		svcMock.
			On("OAuthGetClient", mock.Anything, "client-id").
			Return(&models.OAuthClient{ClientID: "client-id", RedirectURIs: []string{"https://app/cb"}}, nil).
			Once()
		svcMock.
			On("OAuthCreateAuthCode", mock.Anything, "client-id", "user-1", tenant, "https://app/cb", "challenge").
			Return("auth-code", nil).
			Once()

		body := map[string]any{
			"client_id":      "client-id",
			"redirect_uri":   "https://app/cb",
			"code_challenge": "challenge",
			"state":          "xyz",
		}
		data, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/oauth/authorize/callback", strings.NewReader(string(data)))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", bearer)
		rec := httptest.NewRecorder()
		NewRouter(svcMock).ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)

		var resp map[string]string
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
		require.Equal(t, "auth-code", resp["code"])
		require.Equal(t, "xyz", resp["state"])
	})
}

func TestOAuthDynamicRegister(t *testing.T) {
	t.Run("fails when redirect_uris is empty", func(t *testing.T) {
		svcMock := new(servicemock.Service)
		body, _ := json.Marshal(map[string]any{"client_name": "x", "redirect_uris": []string{}})
		req := httptest.NewRequest(http.MethodPost, "/api/oauth/register", strings.NewReader(string(body)))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		NewRouter(svcMock).ServeHTTP(rec, req)
		require.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("rejects non-URL and non-loopback http redirect_uri", func(t *testing.T) {
		// data:/file:/javascript: schemes, malformed URLs, and remote http://
		// hosts (RFC 8252 §8.3 — loopback HTTP is the only allowed plain-HTTP
		// pattern for client registration).
		bad := []string{
			"data:text/html,phish",
			"file:///etc/passwd",
			"javascript:alert(1)",
			"not-a-url",
			"http://attacker.example.com/cb",
			"http://192.168.1.1/cb",
			"ftp://example.com/cb",
		}
		for _, u := range bad {
			svcMock := new(servicemock.Service)
			body, _ := json.Marshal(map[string]any{"client_name": "x", "redirect_uris": []string{u}})
			req := httptest.NewRequest(http.MethodPost, "/api/oauth/register", strings.NewReader(string(body)))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			NewRouter(svcMock).ServeHTTP(rec, req)
			require.Equal(t, http.StatusBadRequest, rec.Code, "redirect_uri %q should be rejected", u)
		}
	})

	t.Run("accepts https and loopback http redirect_uri", func(t *testing.T) {
		good := []string{
			"https://app.example.com/cb",
			"http://localhost/cb",
			"http://localhost:6274/cb",
			"http://127.0.0.1/cb",
			"http://127.0.0.1:8080/cb",
			"http://[::1]:8080/cb",
		}
		for _, u := range good {
			svcMock := new(servicemock.Service)
			svcMock.
				On("OAuthRegisterClient", mock.Anything, "", "x", []string{u}).
				Return(&models.OAuthClient{ClientID: "id", Name: "x", RedirectURIs: []string{u}}, "secret", nil).
				Once()

			body, _ := json.Marshal(map[string]any{"client_name": "x", "redirect_uris": []string{u}})
			req := httptest.NewRequest(http.MethodPost, "/api/oauth/register", strings.NewReader(string(body)))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			NewRouter(svcMock).ServeHTTP(rec, req)
			require.Equal(t, http.StatusCreated, rec.Code, "redirect_uri %q should be accepted", u)
		}
	})

	t.Run("registers a tenantless client (no auth) and returns its credentials", func(t *testing.T) {
		svcMock := new(servicemock.Service)
		svcMock.
			On("OAuthRegisterClient", mock.Anything, "", "x", []string{"http://localhost/cb"}).
			Return(&models.OAuthClient{
				ClientID:     "client-id-1",
				Name:         "x",
				RedirectURIs: []string{"http://localhost/cb"},
			}, "plain-secret", nil).
			Once()

		body, _ := json.Marshal(map[string]any{"client_name": "x", "redirect_uris": []string{"http://localhost/cb"}})
		req := httptest.NewRequest(http.MethodPost, "/api/oauth/register", strings.NewReader(string(body)))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		NewRouter(svcMock).ServeHTTP(rec, req)

		require.Equal(t, http.StatusCreated, rec.Code)

		var resp map[string]any
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
		require.Equal(t, "client-id-1", resp["client_id"])
		require.Equal(t, "plain-secret", resp["client_secret"])
		require.Equal(t, "none", resp["token_endpoint_auth_method"])
	})
}

func TestOAuthRegisterClient(t *testing.T) {
	priv, pub := oauthTestKeys(t)
	svcMock := new(servicemock.Service)
	svcMock.On("PublicKey").Return(pub).Maybe()

	tenant := "00000000-0000-4000-0000-000000000000"
	bearer := oauthBearer(t, priv, authorizer.UserClaims{ID: "user-1", TenantID: tenant})

	cases := []struct {
		description   string
		auth          string
		body          map[string]any
		requiredMocks func()
		expectedCode  int
	}{
		{
			description:   "fails when bearer token is missing",
			body:          map[string]any{"name": "my-client", "redirect_uris": []string{"http://localhost/callback"}},
			requiredMocks: func() {},
			expectedCode:  http.StatusUnauthorized,
		},
		{
			description:   "fails when bearer token is invalid",
			auth:          "Bearer not-a-real-token",
			body:          map[string]any{"name": "my-client", "redirect_uris": []string{"http://localhost/callback"}},
			requiredMocks: func() {},
			expectedCode:  http.StatusUnauthorized,
		},
		{
			description:   "fails when name is too short",
			auth:          bearer,
			body:          map[string]any{"name": "ab", "redirect_uris": []string{"http://localhost/callback"}},
			requiredMocks: func() {},
			expectedCode:  http.StatusBadRequest,
		},
		{
			description:   "fails when redirect_uris is empty",
			auth:          bearer,
			body:          map[string]any{"name": "my-client", "redirect_uris": []string{}},
			requiredMocks: func() {},
			expectedCode:  http.StatusBadRequest,
		},
		{
			description: "succeeds",
			auth:        bearer,
			body:        map[string]any{"name": "my-client", "redirect_uris": []string{"http://localhost/callback"}},
			requiredMocks: func() {
				svcMock.
					On("OAuthRegisterClient", mock.Anything, tenant, "my-client", []string{"http://localhost/callback"}).
					Return(&models.OAuthClient{
						ID:           "id1",
						ClientID:     "client-id-1",
						Name:         "my-client",
						RedirectURIs: []string{"http://localhost/callback"},
					}, "plain-secret", nil).
					Once()
			},
			expectedCode: http.StatusCreated,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/api/oauth/clients", strings.NewReader(string(data)))
			req.Header.Set("Content-Type", "application/json")
			if tc.auth != "" {
				req.Header.Set("Authorization", tc.auth)
			}

			rec := httptest.NewRecorder()
			NewRouter(svcMock).ServeHTTP(rec, req)
			require.Equal(t, tc.expectedCode, rec.Code)
		})
	}
}

func TestOAuthListClients(t *testing.T) {
	priv, pub := oauthTestKeys(t)
	svcMock := new(servicemock.Service)
	svcMock.On("PublicKey").Return(pub).Maybe()

	tenant := "00000000-0000-4000-0000-000000000000"
	bearer := oauthBearer(t, priv, authorizer.UserClaims{ID: "user-1", TenantID: tenant})

	t.Run("fails when bearer token is missing", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/oauth/clients", nil)
		rec := httptest.NewRecorder()
		NewRouter(svcMock).ServeHTTP(rec, req)
		require.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("succeeds", func(t *testing.T) {
		svcMock.
			On("OAuthListClients", mock.Anything, tenant).
			Return([]models.OAuthClient{{ClientID: "client-1"}}, nil).
			Once()

		req := httptest.NewRequest(http.MethodGet, "/api/oauth/clients", nil)
		req.Header.Set("Authorization", bearer)
		rec := httptest.NewRecorder()
		NewRouter(svcMock).ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)
	})
}

func TestOAuthDeleteClient(t *testing.T) {
	priv, pub := oauthTestKeys(t)
	svcMock := new(servicemock.Service)
	svcMock.On("PublicKey").Return(pub).Maybe()

	tenant := "00000000-0000-4000-0000-000000000000"
	bearer := oauthBearer(t, priv, authorizer.UserClaims{ID: "user-1", TenantID: tenant})

	t.Run("fails when bearer token is missing", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/oauth/clients/client-id-1", nil)
		rec := httptest.NewRecorder()
		NewRouter(svcMock).ServeHTTP(rec, req)
		require.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("succeeds", func(t *testing.T) {
		svcMock.
			On("OAuthDeleteClient", mock.Anything, "client-id-1", tenant).
			Return(nil).
			Once()

		req := httptest.NewRequest(http.MethodDelete, "/api/oauth/clients/client-id-1", nil)
		req.Header.Set("Authorization", bearer)
		rec := httptest.NewRecorder()
		NewRouter(svcMock).ServeHTTP(rec, req)
		require.Equal(t, http.StatusNoContent, rec.Code)
	})
}
