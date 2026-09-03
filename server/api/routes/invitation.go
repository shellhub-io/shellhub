package routes

import (
	"context"
	"net/http"

	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/server/api/pkg/gateway"
	"github.com/shellhub-io/shellhub/server/api/pkg/responses"
)

// The registration and invitation routes, relative to the API's base path.
const (
	RegisterUserURL                      = "/register"
	URLResolveInvitation                 = "/invitations/resolve"
	URLUserMembershipInvitationList      = "/users/invitations"
	URLNamespaceMembershipInvitationList = "/namespaces/:tenant/invitations"
	URLGenerateInvitationLink            = "/namespaces/:tenant/invitations/links"
	URLAcceptInvite                      = "/namespaces/:tenant/invitations/accept"
	URLCancelMembershipInvitation        = "/namespaces/:tenant/invitations/:uid"
)

// RegisterUser completes a user account. On the invitation flow the invitee proves email ownership
// via the invite code (sig) and the account is created confirmed, joining the namespace.
func (h *Handler) RegisterUser(c *gateway.Context) error {
	var req requests.RegisterUser

	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	authInfo, err := h.service.RegisterUser(c.Ctx(), req, c.Request().Header.Get("X-Forwarded-Host"), c.Request().Header.Get("X-Forwarded-Proto"))
	if err != nil {
		return err
	}

	if authInfo != nil {
		return c.JSON(http.StatusOK, authInfo)
	}

	return c.NoContent(http.StatusOK)
}

// ResolveInvitation serves what an invite code stands for, so the accept page can route the
// invitee to sign up, log in, or accept. It is reachable unauthenticated: the code is the
// credential.
func (h *Handler) ResolveInvitation(c *gateway.Context) error {
	req := new(requests.ResolveInvitation)

	if err := c.Bind(req); err != nil {
		return err
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	resp, err := h.service.ResolveInvitation(c.Ctx(), req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, resp)
}

// GenerateInvitationLink returns a fresh link for an existing invitation, for an inviter who
// needs to send it again.
func (h *Handler) GenerateInvitationLink(c *gateway.Context) error {
	req := new(requests.GenerateInvitationLink)

	if err := c.Bind(req); err != nil {
		return err
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	link, err := h.service.GenerateInvitationLink(c.Ctx(), req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, map[string]string{"link": link})
}

// AcceptInvite joins the caller to the namespace the invitation names.
func (h *Handler) AcceptInvite(c *gateway.Context) error {
	req := new(requests.AcceptInvite)

	if err := c.Bind(req); err != nil {
		return err
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	if err := h.service.AcceptInvite(c.Ctx(), req); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

// GetUserMembershipInvitationList serves the invitations awaiting the caller.
func (h *Handler) GetUserMembershipInvitationList(ctx context.Context, _ scope.Scope, _ gateway.Actor, req *requests.UserMembershipInvitationList) ([]responses.MembershipInvitation, int, error) {
	return h.service.UserMembershipInvitationList(ctx, req)
}

// GetNamespaceMembershipInvitationList serves the invitations a namespace has outstanding.
func (h *Handler) GetNamespaceMembershipInvitationList(ctx context.Context, _ scope.Scope, _ gateway.Actor, req *requests.NamespaceMembershipInvitationList) ([]responses.MembershipInvitation, int, error) {
	return h.service.NamespaceMembershipInvitationList(ctx, req)
}

// CancelMembershipInvitation withdraws an invitation, invalidating its code.
func (h *Handler) CancelMembershipInvitation(c *gateway.Context) error {
	req := new(requests.CancelMembershipInvitation)

	if err := c.Bind(req); err != nil {
		return err
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	if err := h.service.CancelMembershipInvitation(c.Ctx(), req); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}
