package internalclient

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
)

// sshApprovalAPI defines methods the SSH gateway uses to hold a login open
// while the connecting user decides in the console.
type sshApprovalAPI interface {
	// CreateSSHApproval opens a JIT approval for a held-open SSH login and
	// returns the code the gateway prints in the terminal banner.
	CreateSSHApproval(ctx context.Context, req requests.SSHApprovalCreate) (*models.SSHApprovalCreated, error)

	// GetSSHApprovalStatus asks for the decision, holding the request open until
	// it lands. It returns the state and, once confirmed, the approving user id;
	// a still-pending answer means the API's own wait elapsed.
	GetSSHApprovalStatus(ctx context.Context, code string) (*models.SSHApprovalStatus, error)
}

func (c *client) CreateSSHApproval(ctx context.Context, req requests.SSHApprovalCreate) (*models.SSHApprovalCreated, error) {
	approval := new(models.SSHApprovalCreated)

	resp, err := c.http.
		R().
		SetContext(ctx).
		SetBody(req).
		SetResult(approval).
		Post(c.config.APIBaseURL + "/internal/ssh-approvals")
	if err := HasError(resp, err); err != nil {
		return nil, err
	}

	return approval, nil
}

func (c *client) GetSSHApprovalStatus(ctx context.Context, code string) (*models.SSHApprovalStatus, error) {
	status := new(models.SSHApprovalStatus)

	resp, err := c.http.
		R().
		SetContext(ctx).
		SetPathParam("code", code).
		SetQueryParam("wait", "true").
		SetResult(status).
		Get(c.config.APIBaseURL + "/internal/ssh-approvals/{code}/status")
	if err := HasError(resp, err); err != nil {
		return nil, err
	}

	return status, nil
}
