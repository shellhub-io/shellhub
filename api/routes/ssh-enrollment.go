package routes

import (
	"net/http"

	"github.com/shellhub-io/shellhub/api/pkg/gateway"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
)

const (
	// Create, status and key are internal endpoints the SSH gateway calls;
	// get, confirm and reject require a user token (the enrolling user must be
	// logged in) and are authorized in the service against the target namespace.
	CreateSSHEnrollmentURL    = "/sshid/enrollment"
	GetSSHEnrollmentStatusURL = "/sshid/enrollment/:code/status"
	GetSSHEnrollmentURL       = "/sshid/enrollment/:code"
	ConfirmSSHEnrollmentURL   = "/sshid/enrollment/:code/confirm"
	RejectSSHEnrollmentURL    = "/sshid/enrollment/:code/reject"
	AttachSSHEnrollmentKeyURL = "/sshid/enrollment/:code/key"
)

// AttachSSHEnrollmentKey attaches the presented key to a pending approval so
// the console enrollment page can show its fingerprint and the accept can bind
// it. Internal: called by the SSH gateway once the offered key is known.
func (h *Handler) AttachSSHEnrollmentKey(c gateway.Context) error {
	req := new(requests.SSHEnrollmentKey)
	if err := c.Bind(req); err != nil {
		return err
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	if err := h.service.AttachSSHEnrollmentKey(c.Ctx(), req); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

// CreateSSHEnrollment stores a pending JIT login approval for a held-open SSH
// connection and returns a short-lived code the gateway shows in the terminal.
func (h *Handler) CreateSSHEnrollment(c gateway.Context) error {
	req := new(requests.SSHEnrollmentCreate)
	if err := c.Bind(req); err != nil {
		return err
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	approval, err := h.service.CreateSSHEnrollment(c.Ctx(), req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, approval)
}

// GetSSHEnrollmentStatus reports the decision to the gateway polling it.
func (h *Handler) GetSSHEnrollmentStatus(c gateway.Context) error {
	req := new(requests.SSHEnrollmentStatus)
	if err := c.Bind(req); err != nil {
		return err
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	status, err := h.service.GetSSHEnrollmentStatus(c.Ctx(), req.Code)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, status)
}

// GetSSHEnrollment returns the request details the console renders on the approval
// page so the user sees what they are approving.
func (h *Handler) GetSSHEnrollment(c gateway.Context) error {
	req := new(requests.SSHEnrollmentGet)
	if err := c.Bind(req); err != nil {
		return err
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	approval, err := h.service.GetSSHEnrollment(c.Ctx(), req.Code)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, approval)
}

// ConfirmSSHEnrollment approves a pending login. Authorization against the target's
// namespace happens in the service, since the session may be scoped elsewhere.
func (h *Handler) ConfirmSSHEnrollment(c gateway.Context) error {
	req := new(requests.SSHEnrollmentConfirm)
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

	if err := h.service.ConfirmSSHEnrollment(c.Ctx(), userID, req); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

// RejectSSHEnrollment rejects a pending login.
func (h *Handler) RejectSSHEnrollment(c gateway.Context) error {
	req := new(requests.SSHEnrollmentReject)
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

	if err := h.service.RejectSSHEnrollment(c.Ctx(), userID, req); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}
