package models

import "time"

// OAuthClient represents an OAuth 2.0 client registered in ShellHub.
// Clients use Authorization Code + PKCE flow to obtain JWT tokens.
type OAuthClient struct {
	// ID is the internal unique identifier (UUID).
	ID string `json:"id" bson:"_id"`
	// Name is a human-readable label for the client.
	Name string `json:"name" bson:"name"`
	// ClientID is the public OAuth client identifier (UUID).
	ClientID string `json:"client_id" bson:"client_id"`
	// ClientSecret is the SHA256 hash of the client secret.
	// The plain secret is only returned once at registration time.
	ClientSecret string `json:"-" bson:"client_secret"`
	// TenantID scopes this client to a specific namespace.
	TenantID string `json:"tenant_id" bson:"tenant_id"`
	// RedirectURIs is the allowlist of valid redirect URIs for this client.
	RedirectURIs []string  `json:"redirect_uris" bson:"redirect_uris"`
	CreatedAt    time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" bson:"updated_at"`
}

// OAuthAuthCode is a short-lived authorization code issued during the OAuth
// Authorization Code flow. It is stored in Redis with a 5-minute TTL and
// consumed exactly once during the token exchange.
type OAuthAuthCode struct {
	Code        string `json:"code"`
	ClientID    string `json:"client_id"`
	UserID      string `json:"user_id"`
	TenantID    string `json:"tenant_id"`
	RedirectURI string `json:"redirect_uri"`
	// CodeChallenge is the S256 PKCE challenge sent by the client.
	CodeChallenge string `json:"code_challenge"`
}
