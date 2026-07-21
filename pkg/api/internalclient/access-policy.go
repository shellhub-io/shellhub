package internalclient

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
)

// accessPolicyAPI defines methods the SSH gateway uses to authorize an approved
// identity against a namespace's Access Policies.
type accessPolicyAPI interface {
	// AuthorizeSSHAccess decides whether the user may reach the device as the
	// given login, connecting from sourceIP. It is the identity-mode authorization
	// gate the gateway calls at the ephemeral-key mint point.
	AuthorizeSSHAccess(ctx context.Context, tenant, userID string, device *models.Device, login, sourceIP string) (*models.Decision, error)

	// NamespaceHasAccessPolicies reports whether the namespace has any access
	// policy. The gateway calls it before minting an approval to refuse a login
	// early when no policy could ever grant access.
	NamespaceHasAccessPolicies(ctx context.Context, tenant string) (bool, error)
}

func (c *client) AuthorizeSSHAccess(ctx context.Context, tenant, userID string, device *models.Device, login, sourceIP string) (*models.Decision, error) {
	decision := new(models.Decision)

	resp, err := c.http.
		R().
		SetContext(ctx).
		SetQueryParams(map[string]string{
			"tenant":    tenant,
			"user_id":   userID,
			"device":    device.UID,
			"login":     login,
			"source_ip": sourceIP,
		}).
		SetResult(decision).
		Get(c.config.APIBaseURL + "/internal/access-policies/authorize")
	if err := HasError(resp, err); err != nil {
		return nil, err
	}

	return decision, nil
}

func (c *client) NamespaceHasAccessPolicies(ctx context.Context, tenant string) (bool, error) {
	var out struct {
		Exists bool `json:"exists"`
	}

	resp, err := c.http.
		R().
		SetContext(ctx).
		SetQueryParam("tenant", tenant).
		SetResult(&out).
		Get(c.config.APIBaseURL + "/internal/access-policies/exists")
	if err := HasError(resp, err); err != nil {
		return false, err
	}

	return out.Exists, nil
}
