package routes

import (
	"context"
	"net/http"
	"strconv"

	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/api/responses"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/pkg/authctx"
	"github.com/shellhub-io/shellhub/server/api/pkg/gateway"
	errs "github.com/shellhub-io/shellhub/server/api/routes/errors"
	"github.com/shellhub-io/shellhub/server/api/routes/routetable"
	"github.com/shellhub-io/shellhub/server/api/services"
	log "github.com/sirupsen/logrus"
)

// The namespace and membership routes, relative to the API's base path.
const (
	ListNamespaceURL           = "/namespaces"
	CreateNamespaceURL         = "/namespaces"
	GetNamespaceURL            = "/namespaces/:tenant"
	DeleteNamespaceURL         = "/namespaces/:tenant"
	EditNamespaceURL           = "/namespaces/:tenant"
	ListNamespaceMembersURL    = "/namespaces/:tenant/members"
	LeaveNamespaceURL          = "/namespaces/:tenant/members"
	AddNamespaceMemberURL      = "/namespaces/:tenant/members"
	RemoveNamespaceMemberURL   = "/namespaces/:tenant/members/:uid"
	EditNamespaceMemberURL     = "/namespaces/:tenant/members/:uid"
	EditSessionRecordStatusURL = "/users/security/:tenant"
	UpgradeNamespaceURL        = "/namespaces/:tenant/upgrade"
)

// The path parameter names these routes bind by.
const (
	ParamNamespaceTenant = routetable.TenantParam
)

// GetNamespaceList serves the namespaces the caller belongs to.
func (h *Handler) GetNamespaceList(c *gateway.Context) error {
	req := new(requests.NamespaceList)

	if err := c.Bind(req); err != nil {
		return err
	}

	req.Normalize()

	if err := req.Unmarshal(); err != nil {
		log.WithError(err).WithField("filter", req.Filters.Raw).Warn("failed to decode namespace list filter")

		return errs.NewErrInvalidEntity(map[string]string{"filter": err.Error()})
	}

	if err := query.ValidateFilters(&req.Filters, services.NamespaceFilterFields); err != nil {
		return errs.NewErrInvalidEntity(map[string]string{"filter": err.Error()})
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	namespaces, count, err := h.service.ListNamespaces(c.Ctx(), req)
	if err != nil {
		return err
	}

	c.Response().Header().Set("X-Total-Count", strconv.FormatInt(count, 10))

	return c.JSON(http.StatusOK, responses.NamespacesFromModel(namespaces))
}

// CreateNamespace creates a namespace owned by the caller.
func (h *Handler) CreateNamespace(c *gateway.Context) error {
	req := new(requests.NamespaceCreate)

	if err := c.Bind(req); err != nil {
		return err
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	namespace, err := h.service.CreateNamespace(c.Ctx(), req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, responses.NamespaceFromModel(namespace))
}

// GetNamespace serves the namespace sc is bounded to, by tenant ID. It returns
// NamespaceService.GetNamespace's errors: ErrForbidden when sc was never constructed, and
// ErrNamespaceNotFound when sc is bounded to another namespace or the namespace cannot be resolved.
func (h *Handler) GetNamespace(ctx context.Context, sc scope.Scope, _ gateway.Principal, req *requests.NamespaceGet) (*responses.Namespace, error) {
	return h.getNamespace(ctx, sc, req)
}

// GetNamespaceAsAdmin serves any namespace by tenant ID to an instance administrator, sc spanning
// every namespace. It returns NamespaceService.GetNamespace's errors: ErrForbidden when sc was never
// constructed, and ErrNamespaceNotFound when the namespace cannot be resolved.
func (h *Handler) GetNamespaceAsAdmin(ctx context.Context, sc scope.Scope, _ gateway.Admin, req *requests.NamespaceGet) (*responses.Namespace, error) {
	return h.getNamespace(ctx, sc, req)
}

func (h *Handler) getNamespace(ctx context.Context, sc scope.Scope, req *requests.NamespaceGet) (*responses.Namespace, error) {
	ns, err := h.service.GetNamespace(ctx, sc, req.Tenant)
	if err != nil {
		return nil, err
	}

	return responses.NamespaceFromModel(ns), nil
}

// ListNamespaceMembers serves who belongs to a namespace and in what role.
func (h *Handler) ListNamespaceMembers(ctx context.Context, sc scope.Scope, _ gateway.Principal, req *requests.MemberList) ([]models.MemberView, int64, error) {
	return h.service.ListNamespaceMembers(ctx, sc, req)
}

// DeleteNamespace removes a namespace and everything scoped to it.
func (h *Handler) DeleteNamespace(ctx context.Context, sc scope.Scope, _ gateway.Principal, req *requests.NamespaceDelete) error {
	return h.service.DeleteNamespace(ctx, sc, req.Tenant)
}

// EditNamespace changes a namespace's own attributes, not its membership.
func (h *Handler) EditNamespace(ctx context.Context, sc scope.Scope, _ gateway.Principal, req *requests.NamespaceEdit) (*responses.Namespace, error) {
	ns, err := h.service.EditNamespace(ctx, sc, req)
	if err != nil {
		return nil, err
	}

	return responses.NamespaceFromModel(ns), nil
}

// AddNamespaceMember invites or adds a member in the requested role.
func (h *Handler) AddNamespaceMember(ctx context.Context, sc scope.Scope, _ gateway.Member, req *requests.NamespaceAddMember) (*responses.Namespace, error) {
	ns, err := h.service.AddNamespaceMember(ctx, sc, req)
	if err != nil {
		return nil, err
	}

	return responses.NamespaceFromModel(ns), nil
}

// RemoveNamespaceMember removes another member from the namespace.
func (h *Handler) RemoveNamespaceMember(ctx context.Context, sc scope.Scope, _ gateway.Member, req *requests.NamespaceRemoveMember) (*responses.Namespace, error) {
	ns, err := h.service.RemoveNamespaceMember(ctx, sc, req)
	if err != nil {
		return nil, err
	}

	return responses.NamespaceFromModel(ns), nil
}

// LeaveNamespace removes the caller from the namespace. It is separate from removing a member
// because leaving needs no permission over others. It answers 200 with a replacement token when
// the caller leaves the namespace their own token names, and 204 otherwise, because that token
// keeps working and there is nothing to hand back.
func (h *Handler) LeaveNamespace(c *gateway.Context) error {
	req := new(requests.LeaveNamespace)

	if err := c.Bind(req); err != nil {
		return err
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	if version, ok := authctx.UserTokenVersion(c.Ctx(), req.UserID); ok {
		req.TokenVersion = &version
	}

	req.Impersonated = authctx.UserImpersonated(c.Ctx(), req.UserID)

	res, err := h.service.LeaveNamespace(c.Ctx(), req)
	switch {
	case err != nil:
		return err
	case res != nil:
		return c.JSON(http.StatusOK, res)
	default:
		return c.NoContent(http.StatusNoContent)
	}
}

// EditNamespaceMember changes a member's role.
func (h *Handler) EditNamespaceMember(ctx context.Context, sc scope.Scope, _ gateway.Member, req *requests.NamespaceUpdateMember) error {
	return h.service.UpdateNamespaceMember(ctx, sc, req)
}

// UpgradeNamespace upgrades a legacy namespace to a new one, once and for good. It answers 200 with
// no body, 404 when no namespace carries the tenant, 409 when the namespace is already new, and 403
// to an API key or to a member who may not update the namespace.
func (h *Handler) UpgradeNamespace(ctx context.Context, sc scope.Scope, _ gateway.Member, req *requests.UpgradeNamespace) error {
	return h.service.UpgradeNamespace(ctx, sc, req.Tenant)
}

// EditSessionRecordStatus turns session recording on or off for the namespace.
func (h *Handler) EditSessionRecordStatus(ctx context.Context, sc scope.Scope, _ gateway.Principal, req *requests.SessionEditRecordStatus) error {
	return h.service.EditSessionRecordStatus(ctx, sc, req.SessionRecord, req.Tenant)
}
