package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"time"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
)

const (
	oauthCodeCachePrefix = "oauth:code:"
	oauthCodeTTL         = 5 * time.Minute
)

type OAuthService interface {
	// OAuthRegisterClient creates and persists a new OAuth client for the given tenant.
	// Returns the client with the plain secret (shown only once) and an error, if any.
	OAuthRegisterClient(ctx context.Context, tenantID, name string, redirectURIs []string) (*models.OAuthClient, string, error)

	// OAuthGetClient fetches an OAuth client by its public client_id.
	OAuthGetClient(ctx context.Context, clientID string) (*models.OAuthClient, error)

	// OAuthListClients lists OAuth clients for a given tenant.
	OAuthListClients(ctx context.Context, tenantID string) ([]models.OAuthClient, error)

	// OAuthDeleteClient removes an OAuth client by its internal ID, but only
	// if it belongs to the given tenant. Returns ErrOAuthClientNotFound if the
	// client doesn't exist or belongs to a different tenant (so callers can't
	// probe other tenants' clients via 403/404 differences).
	OAuthDeleteClient(ctx context.Context, id, tenantID string) error

	// OAuthCreateAuthCode generates and caches an authorization code for the given parameters.
	// codeChallenge must be the S256 PKCE challenge (base64url(sha256(verifier))).
	OAuthCreateAuthCode(ctx context.Context, clientID, userID, tenantID, redirectURI, codeChallenge string) (code string, err error)

	// OAuthExchangeCode validates an authorization code and PKCE verifier, then issues a ShellHub JWT.
	// The code is consumed (deleted from cache) on first use.
	OAuthExchangeCode(ctx context.Context, code, clientID, redirectURI, codeVerifier string) (*models.UserAuthResponse, error)
}

func (s *service) OAuthRegisterClient(ctx context.Context, tenantID, name string, redirectURIs []string) (*models.OAuthClient, string, error) {
	plainSecret := uuid.Generate()

	sum := sha256.Sum256([]byte(plainSecret))
	hashedSecret := hex.EncodeToString(sum[:])

	client := &models.OAuthClient{
		ID:           uuid.Generate(),
		Name:         name,
		ClientID:     uuid.Generate(),
		ClientSecret: hashedSecret,
		TenantID:     tenantID,
		RedirectURIs: redirectURIs,
	}

	if _, err := s.store.OAuthClientCreate(ctx, client); err != nil {
		return nil, "", err
	}

	client.ClientSecret = plainSecret

	return client, plainSecret, nil
}

func (s *service) OAuthGetClient(ctx context.Context, clientID string) (*models.OAuthClient, error) {
	client, err := s.store.OAuthClientResolve(ctx, store.OAuthClientClientIDResolver, clientID)
	if err != nil {
		return nil, NewErrNotFound(ErrOAuthClientNotFound, clientID, err)
	}

	return client, nil
}

func (s *service) OAuthListClients(ctx context.Context, tenantID string) ([]models.OAuthClient, error) {
	opts := s.store.Options()
	clients, _, err := s.store.OAuthClientList(ctx, opts.InNamespace(tenantID))
	if err != nil {
		return nil, err
	}

	return clients, nil
}

func (s *service) OAuthDeleteClient(ctx context.Context, id, tenantID string) error {
	client, err := s.store.OAuthClientResolve(ctx, store.OAuthClientIDResolver, id)
	if err != nil {
		return NewErrNotFound(ErrOAuthClientNotFound, id, err)
	}

	if client.TenantID != tenantID {
		return NewErrNotFound(ErrOAuthClientNotFound, id, nil)
	}

	return s.store.OAuthClientDelete(ctx, client)
}

func (s *service) OAuthCreateAuthCode(ctx context.Context, clientID, userID, tenantID, redirectURI, codeChallenge string) (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}

	code := hex.EncodeToString(raw)

	authCode := &models.OAuthAuthCode{
		Code:          code,
		ClientID:      clientID,
		UserID:        userID,
		TenantID:      tenantID,
		RedirectURI:   redirectURI,
		CodeChallenge: codeChallenge,
	}

	if err := s.cache.Set(ctx, oauthCodeCachePrefix+code, authCode, oauthCodeTTL); err != nil {
		return "", err
	}

	return code, nil
}

func (s *service) OAuthExchangeCode(ctx context.Context, code, clientID, redirectURI, codeVerifier string) (*models.UserAuthResponse, error) {
	// RFC 6749 §4.1.2: authorization codes MUST be single-use. Consume the
	// code atomically (Redis GETDEL) so two concurrent requests cannot both
	// observe a hit and issue two JWTs from one code. If the consumption
	// races, exactly one caller gets the value; the others get
	// ErrGetNotFound and are rejected.
	var authCode models.OAuthAuthCode
	if err := s.cache.GetDelete(ctx, oauthCodeCachePrefix+code, &authCode); err != nil {
		return nil, NewErrUnathorized(ErrOAuthCodeInvalid, nil)
	}

	if authCode.Code == "" {
		return nil, NewErrUnathorized(ErrOAuthCodeInvalid, nil)
	}

	if authCode.ClientID != clientID {
		return nil, NewErrUnathorized(ErrOAuthClientIDMismatch, nil)
	}

	if authCode.RedirectURI != redirectURI {
		return nil, NewErrUnathorized(ErrOAuthRedirectURIMismatch, nil)
	}

	if !verifyPKCE(codeVerifier, authCode.CodeChallenge) {
		return nil, NewErrUnathorized(ErrOAuthPKCEInvalid, nil)
	}

	token, err := s.CreateUserToken(ctx, &requests.CreateUserToken{
		UserID:   authCode.UserID,
		TenantID: authCode.TenantID,
	})
	if err != nil {
		return nil, err
	}

	return token, nil
}

// verifyPKCE checks that sha256(verifier) == challenge (base64url, no padding).
// Uses constant-time comparison to avoid timing side-channels and exact byte
// equality per RFC 7636 §4.6 (base64url is case-sensitive).
func verifyPKCE(verifier, challenge string) bool {
	sum := sha256.Sum256([]byte(verifier))
	computed := base64.RawURLEncoding.EncodeToString(sum[:])

	return subtle.ConstantTimeCompare([]byte(computed), []byte(challenge)) == 1
}
