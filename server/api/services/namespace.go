package services

import (
	"context"
	"errors"
	"strings"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	"github.com/shellhub-io/shellhub/server/api/store"
)

// NamespaceFilterFields maps each filter field the namespace list endpoint accepts
// to the set of operators valid for it. The "type" field maps to the "scope" column
// in the database (see namespaceFilterColumns) and only supports equality operators
// because it is an enum column.
var NamespaceFilterFields = query.NewFieldConstraints(map[string][]string{
	"name": {"contains", "eq", "ne"},
	"type": {"eq", "ne"},
})

var namespaceFilterColumns = map[string]string{
	"type": "scope",
}

// NamespaceService owns namespaces themselves: creating, editing and removing them, and the
// settings that apply to everything inside one.
type NamespaceService interface {
	ListNamespaces(ctx context.Context, req *requests.NamespaceList) ([]models.Namespace, int, error)
	CreateNamespace(ctx context.Context, namespace *requests.NamespaceCreate) (*models.Namespace, error)
	GetNamespace(ctx context.Context, tenantID string) (*models.Namespace, error)
	ListNamespaceMembers(ctx context.Context, req *requests.MemberList) ([]models.MemberView, int, error)
	DeleteNamespace(ctx context.Context, tenantID string) error
	EditSessionRecordStatus(ctx context.Context, sessionRecord bool, tenantID string) error
	EditSSHAccessMode(ctx context.Context, sshAccessMode, tenantID string) error
}

// CreateNamespace creates a new namespace.
func (s *service) CreateNamespace(ctx context.Context, req *requests.NamespaceCreate) (*models.Namespace, error) {
	user, err := s.store.UserResolve(ctx, store.UserIDResolver, req.UserID)
	if err != nil || user == nil {
		return nil, NewErrUserNotFound(req.UserID, err)
	}

	if user.MaxNamespaces == 0 {
		return nil, NewErrNamespaceCreationIsForbidden(user.MaxNamespaces, nil)
	} else if user.MaxNamespaces > 0 {
		info, err := s.store.UserGetInfo(ctx, req.UserID)
		switch {
		case err != nil:
			return nil, err
		case len(info.OwnedNamespaces) >= user.MaxNamespaces:
			return nil, NewErrNamespaceLimitReached(user.MaxNamespaces, nil)
		}
	}

	conflictsTarget := &models.NamespaceConflicts{Name: strings.ToLower(req.Name)}
	if _, has, err := s.store.NamespaceConflicts(ctx, conflictsTarget); has || err != nil {
		return nil, NewErrNamespaceDuplicated(err)
	}

	ns := &models.Namespace{
		Name:                 strings.ToLower(req.Name),
		Owner:                user.ID,
		DevicesAcceptedCount: 0,
		DevicesPendingCount:  0,
		DevicesRejectedCount: 0,
		DevicesRemovedCount:  0,
		Members: []models.Member{
			{
				ID:      user.ID,
				Role:    authorizer.RoleOwner,
				AddedAt: clock.Now(),
			},
		},
		Settings: &models.NamespaceSettings{
			SessionRecord:          true,
			ConnectionAnnouncement: envs.AnnouncementFor(envs.CurrentEdition()),
		},
		TenantID: req.TenantID,
		Type:     models.NewDefaultType(),
	}

	if models.IsTypeTeam(req.Type) {
		ns.Type = models.TypeTeam
	} else if models.IsTypePersonal(req.Type) {
		ns.Type = models.TypePersonal
	}

	if req.TenantID == "" {
		ns.TenantID = uuid.Generate()
	}

	if envs.IsCloud() {
		ns.MaxDevices = 3
	} else {
		ns.MaxDevices = -1
	}

	if _, err := s.store.NamespaceCreate(ctx, ns); err != nil {
		if errors.Is(err, store.ErrDuplicate) {
			return nil, NewErrNamespaceDuplicated(err)
		}

		if errors.Is(err, store.ErrNamespaceSingle) {
			return nil, NewErrNamespaceSingle(err)
		}

		return nil, NewErrNamespaceCreateStore(err)
	}

	return ns, nil
}

func (s *service) ListNamespaces(ctx context.Context, req *requests.NamespaceList) ([]models.Namespace, int, error) {
	for i := range req.Filters.Data {
		if p, ok := req.Filters.Data[i].Params.(*query.FilterProperty); ok {
			if col, found := namespaceFilterColumns[p.Name]; found {
				p.Name = col
			}
		}
	}

	if req.UserID == "" && !req.IsAdmin {
		if req.TenantID == "" {
			return []models.Namespace{}, 0, nil
		}

		ns, err := s.store.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, req.TenantID)
		if err != nil {
			return nil, 0, NewErrNamespaceList(err)
		}

		if ns == nil {
			return []models.Namespace{}, 0, nil
		}

		return []models.Namespace{*ns}, 1, nil
	}

	opts := []store.QueryOption{s.store.Options().Match(&req.Filters), s.store.Options().Paginate(&req.Paginator)}
	if req.UserID != "" {
		opts = append(opts, s.store.Options().WithMember(req.UserID))
	}

	namespaces, count, err := s.store.NamespaceList(ctx, opts...)
	if err != nil {
		return nil, 0, NewErrNamespaceList(err)
	}

	return namespaces, count, nil
}

// GetNamespace gets a namespace.
//
// It receives a context, used to "control" the request flow and the tenant ID from models.Namespace.
//
// GetNamespace returns a models.Namespace and an error. When error is not nil, the models.Namespace is nil.
func (s *service) GetNamespace(ctx context.Context, tenantID string) (*models.Namespace, error) {
	namespace, err := s.store.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, tenantID)
	if err != nil || namespace == nil {
		return nil, NewErrNamespaceNotFound(tenantID, err)
	}

	return namespace, nil
}

// ListNamespaceMembers returns the namespace's members as enriched MemberView rows (name,
// username, email, role and a flattened account status), paginated. The tenant is the caller's
// current namespace (X-Tenant-ID). Cloud/enterprise overrides the route to also fold pending
// invitations into the response; core returns only real members.
func (s *service) ListNamespaceMembers(ctx context.Context, req *requests.MemberList) ([]models.MemberView, int, error) {
	if _, err := s.store.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, req.TenantID); err != nil {
		return nil, 0, NewErrNamespaceNotFound(req.TenantID, err)
	}

	sc, err := BoundTo(req.TenantID)
	if err != nil {
		return nil, 0, err
	}

	opts := []store.QueryOption{s.store.Options().Paginate(&req.Paginator)}

	return s.store.NamespaceGetMembers(ctx, sc, opts...)
}

// DeleteNamespace deletes a namespace.
//
// It receives a context, used to "control" the request flow and the tenant ID from models.Namespace.
//
// When cloud and billing is enabled, it will try to delete the namespace's billing information from the billing
// service if it exists.
func (s *service) DeleteNamespace(ctx context.Context, tenantID string) error {
	n, err := s.store.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, tenantID)
	if err != nil {
		return NewErrNamespaceNotFound(tenantID, err)
	}

	ableToReportDeleteNamespace := func(ns *models.Namespace) bool {
		return !ns.Billing.IsNil() && ns.Billing.HasCustomer() && ns.Billing.HasSubscription()
	}

	if envs.IsCloud() && ableToReportDeleteNamespace(n) {
		if err := s.reportBilling(ctx, tenantID, BillingActionNamespaceDelete); err != nil {
			return NewErrBillingReportNamespaceDelete(err)
		}
	}

	if err := fireNamespaceDelete(ctx, n); err != nil {
		return err
	}

	if err := s.store.NamespaceDelete(ctx, n); err != nil {
		if errors.Is(err, store.ErrNamespaceInstanceProtected) {
			return NewErrNamespaceInstanceProtected(err)
		}

		return err
	}

	return nil
}

func (s *service) EditNamespace(ctx context.Context, req *requests.NamespaceEdit) (*models.Namespace, error) {
	namespace, err := s.store.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, req.Tenant)
	if err != nil {
		return nil, NewErrNamespaceNotFound(req.Tenant, err)
	}

	if req.Name != "" && !strings.EqualFold(req.Name, namespace.Name) {
		namespace.Name = strings.ToLower(req.Name)
	}

	if req.Settings.SessionRecord != nil {
		namespace.Settings.SessionRecord = *req.Settings.SessionRecord
	}

	if req.Settings.ConnectionAnnouncement != nil {
		namespace.Settings.ConnectionAnnouncement = *req.Settings.ConnectionAnnouncement
	}

	if err := s.store.NamespaceUpdate(ctx, namespace); err != nil {
		if errors.Is(err, store.ErrDuplicate) {
			return nil, NewErrNamespaceDuplicated(err)
		}

		return nil, err
	}

	return s.store.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, req.Tenant)
}

// EditSessionRecordStatus defines if the sessions will be recorded.
//
// It receives a context, used to "control" the request flow, a boolean to define if the sessions will be recorded and
// the tenant ID from models.Namespace.
//
// This method is deprecated, use [NamespaceService#EditNamespace] instead.
func (s *service) EditSessionRecordStatus(ctx context.Context, sessionRecord bool, tenantID string) error {
	n, err := s.store.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, tenantID)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNoDocuments):
			return NewErrNamespaceNotFound(tenantID, err)
		default:
			return err
		}
	}

	n.Settings.SessionRecord = sessionRecord
	if err := s.store.NamespaceUpdate(ctx, n); err != nil {
		return err
	}

	return nil
}

// EditSSHAccessMode sets the namespace's SSH authorization mode ("legacy" or
// "identity"). Legacy is only reachable by grandfathered namespaces
// (SSHLegacyAllowed); namespaces born identity are refused. Switching to
// "identity" gates SSH logins on browser approval and governs access through
// Access Policies; to avoid a silent lockout on a namespace with no policies,
// the owner-scoped starter policy is seeded on the first switch (see
// seedAccessPolicy).
func (s *service) EditSSHAccessMode(ctx context.Context, sshAccessMode, tenantID string) error {
	n, err := s.store.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, tenantID)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNoDocuments):
			return NewErrNamespaceNotFound(tenantID, err)
		default:
			return err
		}
	}

	if sshAccessMode == models.SSHAccessModeLegacy && !n.Settings.SSHLegacyAllowed {
		return NewErrForbidden(ErrNamespaceLegacyNotAllowed, nil)
	}

	n.Settings.SSHAccessMode = sshAccessMode
	if err := s.store.NamespaceUpdate(ctx, n); err != nil {
		return err
	}

	if sshAccessMode == models.SSHAccessModeIdentity {
		if err := s.seedAccessPolicy(ctx, tenantID, n.Owner); err != nil {
			return err
		}
	}

	return nil
}
