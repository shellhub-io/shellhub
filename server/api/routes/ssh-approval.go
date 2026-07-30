package routes

import (
	"net/http"

	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/server/api/pkg/gateway"
)

const (
	// Create and status are internal endpoints the SSH gateway calls; get,
	// confirm and reject require a user token (the approving user must be logged
	// in) and are authorized in the service against the target namespace.
	GetSSHApprovalURL     = "/ssh-approvals/:code"
	ConfirmSSHApprovalURL = "/ssh-approvals/:code/confirm"
	RejectSSHApprovalURL  = "/ssh-approvals/:code/reject"
)

// GetSSHApproval returns the request details the console renders on the approval
// page so the user sees what they are approving.
func (h *Handler) GetSSHApproval(c gateway.Context) error {
	req := new(requests.SSHApprovalGet)
	if err := c.Bind(req); err != nil {
		return err
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	userID, ok := c.GetID()
	if !ok {
		return c.NoContent(http.StatusUnauthorized)
	}

	approval, err := h.service.GetSSHApproval(c.Ctx(), userID, req.Code)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, approval)
}

// ConfirmSSHApproval approves a pending login. Authorization against the target's
// namespace happens in the service, since the session may be scoped elsewhere.
func (h *Handler) ConfirmSSHApproval(c gateway.Context) error {
	req := new(requests.SSHApprovalConfirm)
	if err := c.Bind(req); err != nil {
		return err
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	userID, ok := c.GetID()
	if !ok {
		return c.NoContent(http.StatusUnauthorized)
	}

	if err := h.service.ConfirmSSHApproval(c.Ctx(), userID, req); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

// RejectSSHApproval rejects a pending login.
func (h *Handler) RejectSSHApproval(c gateway.Context) error {
	req := new(requests.SSHApprovalReject)
	if err := c.Bind(req); err != nil {
		return err
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	userID, ok := c.GetID()
	if !ok {
		return c.NoContent(http.StatusUnauthorized)
	}

	if err := h.service.RejectSSHApproval(c.Ctx(), userID, req); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}
