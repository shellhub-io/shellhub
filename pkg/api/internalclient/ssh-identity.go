package internalclient

import (
	"context"
)

// SSHIdentityResolution is the gateway's view of an identity lookup: whether the
// presented key's fingerprint is enrolled in the namespace and, if so, the bound
// account.
type SSHIdentityResolution struct {
	Found  bool   `json:"found"`
	UserID string `json:"user_id"`
	Name   string `json:"name"`
}

// sshIdentityAPI defines the method the SSH gateway uses to resolve a presented
// public key's fingerprint to a ShellHub identity in the identity access mode.
type sshIdentityAPI interface {
	// ResolveSSHIdentity looks up the fingerprint in the namespace's enrolled
	// identities. A miss (Found=false) means the key must be enrolled before it
	// can be used.
	ResolveSSHIdentity(ctx context.Context, tenant, fingerprint string) (*SSHIdentityResolution, error)
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
