package routes

import (
	"net/http"
	"net/url"

	"github.com/shellhub-io/shellhub/api/pkg/gateway"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/jwttoken"
)

const (
	OAuthMetadataURL        = "/.well-known/oauth-authorization-server"
	OAuthAuthorizeURL       = "/oauth/authorize"
	OAuthTokenURL           = "/oauth/token" //nolint:gosec // URL path, not a credential
	OAuthClientsURL         = "/oauth/clients"
	OAuthClientURL          = "/oauth/clients/:id"
	OAuthDynamicRegisterURL = "/oauth/register"
)

// OAuthMetadata returns RFC 8414 Authorization Server Metadata.
func (h *Handler) OAuthMetadata(c gateway.Context) error {
	r := c.Request()

	scheme := r.Header.Get("X-Forwarded-Proto")
	if scheme == "" {
		scheme = "https"
		if r.TLS == nil {
			scheme = "http"
		}
	}

	host := r.Header.Get("X-Forwarded-Host")
	if host == "" {
		host = r.Host
	}

	port := r.Header.Get("X-Forwarded-Port")
	if port != "" && port != "80" && port != "443" {
		host = host + ":" + port
	}

	baseURL := scheme + "://" + host

	return c.JSON(http.StatusOK, map[string]any{
		"issuer":                                baseURL,
		"authorization_endpoint":                baseURL + "/api/oauth/authorize",
		"token_endpoint":                        baseURL + "/api/oauth/token",
		"registration_endpoint":                 baseURL + "/api/oauth/register",
		"response_types_supported":              []string{"code"},
		"grant_types_supported":                 []string{"authorization_code"},
		"code_challenge_methods_supported":      []string{"S256"},
		"token_endpoint_auth_methods_supported": []string{"none"},
	})
}

// OAuthAuthorize handles GET /oauth/authorize.
// It validates the client and stores the auth code context, then redirects to the UI login page.
func (h *Handler) OAuthAuthorize(c gateway.Context) error {
	clientID := c.QueryParam("client_id")
	redirectURI := c.QueryParam("redirect_uri")
	codeChallenge := c.QueryParam("code_challenge")
	codeChallengeMethod := c.QueryParam("code_challenge_method")
	state := c.QueryParam("state")

	if codeChallengeMethod != "S256" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error":             "invalid_request",
			"error_description": "code_challenge_method must be S256",
		})
	}

	client, err := h.service.OAuthGetClient(c.Ctx(), clientID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error":             "invalid_client",
			"error_description": "unknown client_id",
		})
	}

	if !isAllowedRedirectURI(client.RedirectURIs, redirectURI) {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error":             "invalid_request",
			"error_description": "redirect_uri not allowed",
		})
	}

	// The login callback retrieves these from the query string after the user
	// authenticates. Use url.Values to safely encode user-supplied values
	// (redirect_uri, state) so '&', '#', and other reserved chars cannot inject
	// or override other parameters.
	params := url.Values{}
	params.Set("oauth_client_id", clientID)
	params.Set("oauth_redirect_uri", redirectURI)
	params.Set("oauth_code_challenge", codeChallenge)
	params.Set("oauth_state", state)

	return c.Redirect(http.StatusFound, "/login?"+params.Encode())
}

// oauthUserFromBearer validates the request's Authorization header against the
// API public key and returns the caller's claims. The /api/oauth/* gateway
// block has auth_request off, so each handler that needs identity must call
// this helper instead of trusting client-supplied X-* headers.
func (h *Handler) oauthUserFromBearer(c gateway.Context) (*authorizer.UserClaims, bool) {
	claims, err := jwttoken.ClaimsFromBearerToken(h.service.PublicKey(), c.Request().Header.Get("Authorization"))
	if err != nil {
		return nil, false
	}

	userClaims, ok := claims.(*authorizer.UserClaims)
	if !ok {
		return nil, false
	}

	return userClaims, true
}

// OAuthCallback handles POST /oauth/authorize/callback.
// Called by the ShellHub UI after the user logs in; it issues the authorization code.
func (h *Handler) OAuthCallback(c gateway.Context) error {
	type req struct {
		ClientID      string `json:"client_id"`
		RedirectURI   string `json:"redirect_uri"`
		CodeChallenge string `json:"code_challenge"`
		State         string `json:"state"`
	}

	var r req
	if err := c.Bind(&r); err != nil {
		return err
	}

	user, ok := h.oauthUserFromBearer(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	client, err := h.service.OAuthGetClient(c.Ctx(), r.ClientID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_client"})
	}

	if !isAllowedRedirectURI(client.RedirectURIs, r.RedirectURI) {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_request"})
	}

	code, err := h.service.OAuthCreateAuthCode(c.Ctx(), r.ClientID, user.ID, user.TenantID, r.RedirectURI, r.CodeChallenge)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, map[string]string{
		"code":  code,
		"state": r.State,
	})
}

// OAuthToken handles POST /oauth/token — exchanges authorization code for JWT.
func (h *Handler) OAuthToken(c gateway.Context) error {
	grantType := c.FormValue("grant_type")
	if grantType != "authorization_code" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error":             "unsupported_grant_type",
			"error_description": "only authorization_code is supported",
		})
	}

	code := c.FormValue("code")
	clientID := c.FormValue("client_id")
	redirectURI := c.FormValue("redirect_uri")
	codeVerifier := c.FormValue("code_verifier")

	token, err := h.service.OAuthExchangeCode(c.Ctx(), code, clientID, redirectURI, codeVerifier)
	if err != nil {
		// Don't surface internal error chains (which may include user IDs,
		// DB error messages, etc.) on the OAuth wire — return a generic
		// invalid_grant. The actual cause is logged by the service.
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error":             "invalid_grant",
			"error_description": "authorization code is invalid, expired, or has already been used",
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"access_token": token.Token,
		"token_type":   "Bearer",
		"expires_in":   72 * 3600,
		"scope":        "openid",
	})
}

// OAuthDynamicRegister handles POST /oauth/register — RFC 7591 Dynamic Client Registration.
// No authentication required; used by MCP clients to self-register.
func (h *Handler) OAuthDynamicRegister(c gateway.Context) error {
	type req struct {
		ClientName   string   `json:"client_name"`
		RedirectURIs []string `json:"redirect_uris" validate:"required,min=1,dive,url"`
	}

	var r req
	if err := c.Bind(&r); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_request"})
	}

	// Same baseline validation as the authenticated /oauth/clients endpoint:
	// each redirect_uri must parse as a URL.
	if err := c.Validate(&r); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_redirect_uri"})
	}

	if !areAllowedRedirectURISchemes(r.RedirectURIs) {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_redirect_uri"})
	}

	name := r.ClientName
	if name == "" {
		name = "mcp-client"
	}

	// Dynamic clients have no tenant — tenantID resolved at token exchange from the auth code.
	client, plainSecret, err := h.service.OAuthRegisterClient(c.Ctx(), "", name, r.RedirectURIs)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, map[string]any{
		"client_id":                  client.ClientID,
		"client_secret":              plainSecret,
		"client_name":                client.Name,
		"redirect_uris":              client.RedirectURIs,
		"grant_types":                []string{"authorization_code"},
		"response_types":             []string{"code"},
		"token_endpoint_auth_method": "none",
	})
}

// OAuthRegisterClient handles POST /oauth/clients.
func (h *Handler) OAuthRegisterClient(c gateway.Context) error {
	type req struct {
		Name         string   `json:"name" validate:"required,min=3,max=64"`
		RedirectURIs []string `json:"redirect_uris" validate:"required,min=1,dive,url"`
	}

	var r req
	if err := c.Bind(&r); err != nil {
		return err
	}

	if err := c.Validate(&r); err != nil {
		return err
	}

	if !areAllowedRedirectURISchemes(r.RedirectURIs) {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_redirect_uri"})
	}

	user, ok := h.oauthUserFromBearer(c)
	if !ok || user.TenantID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	client, plainSecret, err := h.service.OAuthRegisterClient(c.Ctx(), user.TenantID, r.Name, r.RedirectURIs)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, map[string]any{
		"id":            client.ID,
		"client_id":     client.ClientID,
		"client_secret": plainSecret,
		"name":          client.Name,
		"redirect_uris": client.RedirectURIs,
	})
}

// OAuthListClients handles GET /oauth/clients.
func (h *Handler) OAuthListClients(c gateway.Context) error {
	user, ok := h.oauthUserFromBearer(c)
	if !ok || user.TenantID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	clients, err := h.service.OAuthListClients(c.Ctx(), user.TenantID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, clients)
}

// OAuthDeleteClient handles DELETE /oauth/clients/:id.
func (h *Handler) OAuthDeleteClient(c gateway.Context) error {
	user, ok := h.oauthUserFromBearer(c)
	if !ok || user.TenantID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	if err := h.service.OAuthDeleteClient(c.Ctx(), c.Param("id"), user.TenantID); err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}

// areAllowedRedirectURISchemes enforces the redirect_uri scheme allowlist for
// client registration. Only HTTPS is allowed for arbitrary hosts; HTTP is
// only permitted for loopback addresses (RFC 8252 §8.3 — loopback HTTP is
// the standard pattern for native CLI/MCP clients). Anything else (data:,
// file:, javascript:, http://attacker.com, ...) would let a registrant
// hijack the authorization flow or escape the OAuth response handling.
func areAllowedRedirectURISchemes(uris []string) bool {
	for _, u := range uris {
		parsed, err := url.Parse(u)
		if err != nil {
			return false
		}

		switch parsed.Scheme {
		case "https":
			// Always allowed — any host.
		case "http":
			host := parsed.Hostname()
			if host != "localhost" && host != "127.0.0.1" && host != "::1" {
				return false
			}
		default:
			return false
		}
	}

	return true
}

func isAllowedRedirectURI(allowed []string, target string) bool {
	for _, u := range allowed {
		if u == target {
			return true
		}
	}

	return false
}
