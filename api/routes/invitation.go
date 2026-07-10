package routes

import (
	"net/http"
	"strconv"

	"github.com/shellhub-io/shellhub/api/pkg/gateway"
	"github.com/shellhub-io/shellhub/api/services"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/errors"
	log "github.com/sirupsen/logrus"
)

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
func (h *Handler) RegisterUser(c gateway.Context) error {
	var req requests.RegisterUser

	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	authInfo, conflictFields, err := h.service.RegisterUser(c.Ctx(), req, c.Request().Header.Get("X-Forwarded-Host"))
	if err != nil {
		// The UI uses the conflicting fields to tell invalid from duplicated.
		var e errors.Error
		if ok := errors.As(err, &e); !ok {
			return err
		}

		switch e.Code {
		case services.ErrCodeInvalid:
			return c.JSON(http.StatusBadRequest, conflictFields)
		case services.ErrCodeDuplicated:
			return c.JSON(http.StatusConflict, conflictFields)
		default:
			return err
		}
	}

	if authInfo != nil {
		return c.JSON(http.StatusOK, authInfo)
	}

	return c.NoContent(http.StatusOK)
}

func (h *Handler) ResolveInvitation(c gateway.Context) error {
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

func (h *Handler) GenerateInvitationLink(c gateway.Context) error {
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

func (h *Handler) AcceptInvite(c gateway.Context) error {
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

func (h *Handler) GetUserMembershipInvitationList(c gateway.Context) error {
	req := new(requests.UserMembershipInvitationList)

	if err := c.Bind(req); err != nil {
		return err
	}

	req.Paginator.Normalize()
	req.Sorter.Normalize()

	if err := req.Filters.Unmarshal(); err != nil { //nolint:staticcheck
		log.WithError(err).WithField("filter", req.Filters.Raw).Warn("failed to decode user membership invitation list filter")

		return c.NoContent(http.StatusBadRequest)
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	invitations, count, err := h.service.UserMembershipInvitationList(c.Ctx(), req)
	if err != nil {
		return err
	}

	c.Response().Header().Set("X-Total-Count", strconv.FormatInt(count, 10))

	return c.JSON(http.StatusOK, invitations)
}

func (h *Handler) GetNamespaceMembershipInvitationList(c gateway.Context) error {
	req := new(requests.NamespaceMembershipInvitationList)

	if err := c.Bind(req); err != nil {
		return err
	}

	req.Paginator.Normalize()
	req.Sorter.Normalize()

	if err := req.Filters.Unmarshal(); err != nil { //nolint:staticcheck
		log.WithError(err).WithField("filter", req.Filters.Raw).Warn("failed to decode namespace membership invitation list filter")

		return c.NoContent(http.StatusBadRequest)
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	invitations, count, err := h.service.NamespaceMembershipInvitationList(c.Ctx(), req)
	if err != nil {
		return err
	}

	c.Response().Header().Set("X-Total-Count", strconv.FormatInt(count, 10))

	return c.JSON(http.StatusOK, invitations)
}

func (h *Handler) CancelMembershipInvitation(c gateway.Context) error {
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
