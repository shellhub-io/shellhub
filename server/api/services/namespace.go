package services

import (
	"context"
	"errors"
	"strings"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	"github.com/shellhub-io/shellhub/server/api/store"
)

// NamespaceQuery is the query contract the namespace list accepts. The "type" field maps to the
// "scope" column in the database (see namespaceFilterColumns) and only supports equality operators
// because it is an enum column. It declares the values it accepts for the same reason, and declares
// them as the types the API names rather than the scopes the column stores, because the rewrite to
// the column happens after validation. The list takes no sort.
var NamespaceQuery = query.Contract{
	Filter: query.NewFieldConstraints(map[string][]string{
		"name": {"contains", "eq", "ne"},
		"type": {"eq", "ne"},
	}).WithValues(map[string][]string{
		"type": {
			string(models.TypePersonal),
			string(models.TypeTeam),
		},
	}),
}

// MemberQuery is the query contract the namespace member list accepts, which is nothing: the list
// pages through a namespace's members and offers neither a filter nor a sort.
var MemberQuery = query.Contract{}

var namespaceFilterColumns = map[string]string{
	"type": "scope",
}

// NamespaceService owns namespaces themselves: creating, editing and removing them, and the
// settings that apply to everything inside one.
type NamespaceService interface {
	ListNamespaces(ctx context.Context, req *requests.NamespaceList) ([]models.Namespace, int64, error)
	CreateNamespace(ctx context.Context, namespace *requests.NamespaceCreate) (*models.Namespace, error)
	GetNamespace(ctx context.Context, tenantID string) (*models.Namespace, error)
	ListNamespaceMembers(ctx context.Context, req *requests.MemberList) ([]models.MemberView, int64, error)
	DeleteNamespace(ctx context.Context, tenantID string) error
	EditSessionRecordStatus(ctx context.Context, sessionRecord bool, tenantID string) error
	// UpgradeNamespace upgrades a legacy namespace to a new one, once and for good. The
	// flag and the owner's starter Access Policy, skipped when the namespace already has one, are
	// written in one transaction, so a namespace never reaches the new model with no policy to
	// authorize anyone. It returns ErrNamespaceNotFound when no namespace carries the tenant,
	// ErrNamespaceNotLegacy when the namespace is already new, and the store's error otherwise,
	// with the upgrade rolled back when the failure comes inside the transaction.
	UpgradeNamespace(ctx context.Context, tenantID string) error
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
				AddedAt: s.clock.Now(),
			},
		},
		Settings: &models.NamespaceSettings{
			SessionRecord:          true,
			ConnectionAnnouncement: envs.AnnouncementFor(s.edition),
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

	if s.edition.IsCloud() {
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

func (s *service) ListNamespaces(ctx context.Context, req *requests.NamespaceList) ([]models.Namespace, int64, error) {
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
func (s *service) ListNamespaceMembers(ctx context.Context, req *requests.MemberList) ([]models.MemberView, int64, error) {
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
//
// After the namespace is gone it ends every device it held as a device removal does, evicting its cached
// authorization and closing its tunnel when this process holds it, or on the tunnel's next beat otherwise. A
// failed eviction is logged and does not fail the call. It returns ErrNamespaceNotFound, ErrBillingReportNamespaceDelete when
// billing refuses the deletion, a namespace delete hook's error, ErrNamespaceInstanceProtected for the instance's
// namespace, ErrForbidden for a tenant ID that is not a valid scope, and the store's error when the namespace's
// devices cannot be listed or the namespace cannot be deleted, evicting nothing on any of them. The devices are
// listed before billing hears of the deletion and before the delete hooks run, so a failed listing leaves no
// side effect behind. ctx bounds the listing, the billing report, the hooks, the deletion and the evictions, so a
// ctx cancelled after the deletion leaves the cached authorizations to expire on their own.
func (s *service) DeleteNamespace(ctx context.Context, tenantID string) error {
	n, err := s.store.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, tenantID)
	if err != nil {
		return NewErrNamespaceNotFound(tenantID, err)
	}

	err = DeleteNamespacesEndingDevices(ctx, s.store, s.cache, []string{n.TenantID}, func(ctx context.Context) error {
		return s.deleteNamespace(ctx, n)
	})
	if errors.Is(err, store.ErrNamespaceInstanceProtected) {
		return NewErrNamespaceInstanceProtected(err)
	}

	return err
}

func (s *service) deleteNamespace(ctx context.Context, n *models.Namespace) error {
	ableToReportDeleteNamespace := func(ns *models.Namespace) bool {
		return !ns.Billing.IsNil() && ns.Billing.HasCustomer() && ns.Billing.HasSubscription()
	}

	if s.edition.IsCloud() && ableToReportDeleteNamespace(n) {
		if err := s.reportBilling(ctx, n.TenantID, BillingActionNamespaceDelete); err != nil {
			return NewErrBillingReportNamespaceDelete(err)
		}
	}

	if err := fireNamespaceDelete(ctx, n); err != nil {
		return err
	}

	return s.store.NamespaceDelete(ctx, n)
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

func (s *service) UpgradeNamespace(ctx context.Context, tenantID string) error {
	n, err := s.store.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, tenantID)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNoDocuments):
			return NewErrNamespaceNotFound(tenantID, err)
		default:
			return err
		}
	}

	return s.store.WithTransaction(ctx, func(ctx context.Context) error {
		if err := s.store.NamespaceUpgrade(ctx, tenantID); err != nil {
			if errors.Is(err, store.ErrNamespaceNotLegacy) {
				return NewErrNamespaceNotLegacy(tenantID, err)
			}

			return err
		}

		return s.seedAccessPolicy(ctx, tenantID, n.Owner)
	})
}
