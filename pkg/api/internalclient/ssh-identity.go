package internalclient

import (
	"context"
	"time"
)

// SSHIdentityResolution is the gateway's view of an identity lookup: whether the
// presented key's fingerprint is enrolled in the namespace and, if so, the bound
// account.
type SSHIdentityResolution struct {
	Found  bool   `json:"found"`
	UserID string `json:"user_id"`
	Name   string `json:"name"`
	// LastReauthAt is when the identity last completed a re-authentication, used
	// by the gateway to honor a policy's reauth_period freshness window. Nil when
	// it never re-authed.
	LastReauthAt *time.Time `json:"last_reauth_at"`
	// Active is the key's derived liveness (not expired, not consumed), computed
	// server-side so the gateway never re-derives expiry. A Found-but-inactive
	// key is a dead service-account key: reject the connection, don't enroll.
	Active bool `json:"active"`
	// SingleUse tells the gateway to burn the key after the session establishes.
	SingleUse bool `json:"single_use"`
}

// sshIdentityAPI defines the methods the SSH gateway uses to resolve, and burn,
// a presented public key's fingerprint in the identity access mode.
type sshIdentityAPI interface {
	// ResolveSSHIdentity looks up the fingerprint in the namespace's enrolled
	// identities. A miss (Found=false) means the key must be enrolled before it
	// can be used.
	ResolveSSHIdentity(ctx context.Context, tenant, fingerprint string) (*SSHIdentityResolution, error)
	// ConsumeSSHIdentity atomically burns a single-use identity once its session
	// is established. It returns true when this call won the burn; false means the
	// key was already consumed (a lost race) and the session must be denied.
	ConsumeSSHIdentity(ctx context.Context, tenant, fingerprint string) (bool, error)
}

func (c *client) ResolveSSHIdentity(ctx context.Context, tenant, fingerprint string) (*SSHIdentityResolution, error) {
	resolution := new(SSHIdentityResolution)

	resp, err := c.http.
		R().
		SetContext(ctx).
		SetQueryParams(map[string]string{
			"tenant":      tenant,
			"fingerprint": fingerprint,
		}).
		SetResult(resolution).
		Get(c.config.APIBaseURL + "/internal/ssh-identities/resolve")
	if err := HasError(resp, err); err != nil {
		return nil, err
	}

	return resolution, nil
}

func (c *client) ConsumeSSHIdentity(ctx context.Context, tenant, fingerprint string) (bool, error) {
	result := new(struct {
		Consumed bool `json:"consumed"`
	})

	resp, err := c.http.
		R().
		SetContext(ctx).
		SetQueryParams(map[string]string{
			"tenant":      tenant,
			"fingerprint": fingerprint,
		}).
		SetResult(result).
		Post(c.config.APIBaseURL + "/internal/ssh-identities/consume")
	if err := HasError(resp, err); err != nil {
		return false, err
	}

	return result.Consumed, nil
}
