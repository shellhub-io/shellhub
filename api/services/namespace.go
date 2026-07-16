package services

import (
	"context"
	"errors"
	"strings"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
)

// NamespaceFilterFields maps each filter field the namespace list endpoint accepts
// to the set of operators valid for it. The "type" field maps to the "scope" column
// in the database (see namespaceFilterColumns) and only supports equality operators
// because it is an enum column.
var NamespaceFilterFields = query.NewFieldConstraints(map[string][]string{
	"name": {"contains", "eq", "ne"},
	"type": {"eq", "ne"},
})

// namespaceFilterColumns maps API-level field names to their actual database column
// names where the two differ. It is used by the store layer to translate filter
// properties before constructing SQL queries.
var namespaceFilterColumns = map[string]string{
	"type": "scope",
}

type NamespaceService interface {
	ListNamespaces(ctx context.Context, req *requests.NamespaceList) ([]models.Namespace, int, error)
	CreateNamespace(ctx context.Context, namespace *requests.NamespaceCreate) (*models.Namespace, error)
	GetNamespace(ctx context.Context, tenantID string) (*models.Namespace, error)
	ListNamespaceMembers(ctx context.Context, req *requests.MemberList) ([]models.MemberView, int, error)
	DeleteNamespace(ctx context.Context, tenantID string) error
	EditSessionRecordStatus(ctx context.Context, sessionRecord bool, tenantID string) error
}

// CreateNamespace creates a new namespace.
func (s *service) CreateNamespace(ctx context.Context, req *requests.NamespaceCreate) (*models.Namespace, error) {
	user, err := s.store.UserResolve(ctx, store.UserIDResolver, req.UserID)
	if err != nil || user == nil {
		return nil, NewErrUserNotFound(req.UserID, err)
	}

	// When MaxNamespaces is less than zero, it means that the user has no limit
	// of namespaces. If the value is zero, it means he has no right to create a new namespace
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
			ConnectionAnnouncement: "",
		},
		TenantID: req.TenantID,
		Type:     models.NewDefaultType(),
	}

	if envs.IsCommunity() {
		ns.Settings.ConnectionAnnouncement = models.DefaultAnnouncementMessage
	}

	if models.IsTypeTeam(req.Type) {
		ns.Type = models.TypeTeam
	} else if models.IsTypePersonal(req.Type) {
		ns.Type = models.TypePersonal
	}

	if req.TenantID == "" {
		ns.TenantID = uuid.Generate()
	}

	// Set limits according to ShellHub instance type
	if envs.IsCloud() {
		// cloud free plan is limited only by the max of devices
		ns.MaxDevices = 3
	} else {
		// we don't set limits on enterprise and community instances
		ns.MaxDevices = -1
	}

	// The NamespaceConflicts pre-check above is the fast path; store.ErrDuplicate
	// here means a concurrent insert raced past it. Map it to ErrNamespaceDuplicated
	// so callers get a consistent duplicate signal regardless of timing.
	if _, err := s.store.NamespaceCreate(ctx, ns); err != nil {
		if errors.Is(err, store.ErrDuplicate) {
			return nil, NewErrNamespaceDuplicated(err)
		}

		// Defense in depth: the route is dropped in Community, but if the edition env and the
		// store binding ever disagree (e.g. an Enterprise env running against a bound CE store),
		// surface the single-namespace refusal as a clean conflict instead of a 500.
		if errors.Is(err, store.ErrNamespaceSingle) {
			return nil, NewErrNamespaceSingle(err)
		}

		return nil, NewErrNamespaceCreateStore(err)
	}

	// The namespace's legacy install key is created by the store at namespace creation, so every
	// creation path (here, setup, the CLI, cloud/enterprise) gets it uniformly.

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

	// When the caller has no user ID and is not a system admin (e.g.
	// authenticated via API key), the listing is scoped to the caller's
	// tenant. Otherwise the caller could enumerate namespaces across tenants.
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

	opts := []store.QueryOption{s.store.Options().Paginate(&req.Paginator)}

	return s.store.NamespaceGetMembers(ctx, req.TenantID, opts...)
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
		return !ns.Billing.IsNil() && ns.Billing.HasCutomer() && ns.Billing.HasSubscription()
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
		// The instance is bound to this namespace (single-namespace Community deployment); the
		// FK's ON DELETE RESTRICT refuses it. Surface it as a 409 instead of a 500.
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

	// NamespaceUpdate returns store.ErrDuplicate when the new name collides with an
	// existing namespace. Map it to ErrNamespaceDuplicated so callers get a
	// consistent duplicate signal regardless of timing.
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
	if err := s.store.NamespaceUpdate(ctx, n); err != nil { // nolint:revive
		return err
	}

	return nil
}
