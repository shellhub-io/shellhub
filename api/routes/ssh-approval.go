package routes

import (
	"net/http"

	"github.com/shellhub-io/shellhub/api/pkg/gateway"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
)

const (
	// Create and status are internal endpoints the SSH gateway calls; get,
	// confirm and reject require a user token (the approving user must be logged
	// in) and are authorized in the service against the target namespace.
	CreateSSHApprovalURL    = "/ssh-approvals"
	GetSSHApprovalStatusURL = "/ssh-approvals/:code/status"
	GetSSHApprovalURL       = "/ssh-approvals/:code"
	ConfirmSSHApprovalURL   = "/ssh-approvals/:code/confirm"
	RejectSSHApprovalURL    = "/ssh-approvals/:code/reject"
)

// CreateSSHApproval opens a pending JIT login approval for a held-open SSH
// connection and returns a short-lived code the gateway shows in the terminal.
func (h *Handler) CreateSSHApproval(c gateway.Context) error {
	req := new(requests.SSHApprovalCreate)
	if err := c.Bind(req); err != nil {
		return err
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	approval, err := h.service.CreateSSHApproval(c.Ctx(), req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, approval)
}

// GetSSHApprovalStatus reports the decision to the gateway waiting on it. With
// ?wait it answers only once the login is decided, or once its own wait elapses.
func (h *Handler) GetSSHApprovalStatus(c gateway.Context) error {
	req := new(requests.SSHApprovalStatus)
	if err := c.Bind(req); err != nil {
		return err
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	status, err := h.service.GetSSHApprovalStatus(c.Ctx(), req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, status)
}

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
