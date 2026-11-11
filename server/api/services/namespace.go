package services

import (
	"context"
	"errors"
	"strings"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/store"
)

// NamespaceFilterFields maps each filter field the namespace list endpoint accepts
// to the set of operators valid for it. The "type" field maps to the "scope" column
// in the database (see namespaceFilterColumns) and only supports equality operators
// because it is an enum column. It declares the values it accepts for the same reason,
// and declares them as the types the API names rather than the scopes the column stores,
// because the rewrite to the column happens after validation.
var NamespaceFilterFields = query.NewFieldConstraints(map[string][]string{
	"name": {"contains", "eq", "ne"},
	"type": {"eq", "ne"},
}).WithValues(map[string][]string{
	"type": {
		string(models.TypePersonal),
		string(models.TypeTeam),
	},
})

var namespaceFilterColumns = map[string]string{
	"type": "scope",
}

// NamespaceService owns namespaces themselves: creating, editing and removing them, and the
// settings that apply to everything inside one.
type NamespaceService interface {
	ListNamespaces(ctx context.Context, req *requests.NamespaceList) ([]models.Namespace, int64, error)
	// CreateNamespace creates a namespace owned by the requesting user, under a tenant ID the store
	// issues. For a user with a namespace limit, concurrent calls count and insert one at a time,
	// so the limit holds under parallel requests, and a call waiting on another returns ctx's
	// error when ctx ends first. It
	// returns ErrUserNotFound when the user cannot be resolved, whatever the store's error;
	// ErrNamespaceCreationIsForbidden when the user may own no namespace; ErrNamespaceLimitReached
	// when the user already owns as many as the limit allows; ErrNamespaceDuplicated when the name
	// is taken or the name check fails; ErrNamespaceSingle when the instance is bound to a single
	// namespace; ErrNamespaceCreateStore when the namespace cannot be stored; and the store's
	// error when the transaction cannot begin or commit, the owner lock cannot be taken or the
	// owned namespaces cannot be read. Any failure leaves nothing written.
	CreateNamespace(ctx context.Context, namespace *requests.NamespaceCreate) (*models.Namespace, error)
	// GetNamespace returns the namespace carrying tenantID. It returns ErrForbidden when sc was never
	// constructed, and ErrNamespaceNotFound when sc is bounded to another namespace or the namespace
	// carrying tenantID cannot be resolved, whatever the store's error.
	GetNamespace(ctx context.Context, sc scope.Scope, tenantID string) (*models.Namespace, error)
	// ListNamespaceMembers returns a page of the members of the namespace carrying req.TenantID. It
	// returns ErrForbidden when sc was never constructed, ErrNamespaceNotFound when sc is bounded to
	// another namespace, req.TenantID is empty, or the namespace carrying it cannot be resolved,
	// whatever the store's error; and the store's error when the members cannot be read.
	ListNamespaceMembers(ctx context.Context, sc scope.Scope, req *requests.MemberList) ([]models.MemberView, int64, error)
	// DeleteNamespace deletes the namespace carrying tenantID. When cloud and billing is enabled, it
	// reports the deletion to billing first. After the namespace is gone it ends every device it held
	// as a device removal does, evicting its cached authorization and closing its tunnel when this
	// process holds it, or on the tunnel's next beat otherwise; a failed eviction is logged and does
	// not fail the call.
	//
	// It returns ErrForbidden when sc was never constructed; ErrNamespaceNotFound when sc is bounded
	// to another namespace or the namespace carrying tenantID cannot be resolved, whatever the store's
	// error; ErrBillingReportNamespaceDelete when billing refuses the deletion; a namespace delete
	// hook's error; ErrNamespaceInstanceProtected for the instance's namespace; and the store's error
	// when the namespace's devices cannot be listed or the namespace cannot be deleted. None of them
	// evicts anything. The devices are listed before billing hears of the deletion and before the
	// delete hooks run, so a failed listing leaves no side effect behind. ctx bounds the listing, the
	// billing report, the hooks, the deletion and the evictions, so a ctx cancelled after the deletion
	// leaves the cached authorizations to expire on their own.
	DeleteNamespace(ctx context.Context, sc scope.Scope, tenantID string) error
	// EditSessionRecordStatus turns session recording on or off in the namespace carrying tenantID. It
	// returns ErrForbidden when sc was never constructed, ErrNamespaceNotFound when sc is bounded to
	// another namespace or no namespace carries tenantID, and the store's error otherwise.
	EditSessionRecordStatus(ctx context.Context, sc scope.Scope, sessionRecord bool, tenantID string) error
	// UpgradeNamespace upgrades a legacy namespace to a new one, once and for good. The flag and the
	// owner's starter Access Policy, skipped when the namespace already has one, are written in one
	// transaction, so a namespace never reaches the new model with no policy to authorize anyone. It
	// returns ErrNamespaceNotFound when no namespace carries the tenant, or sc is bounded to another
	// namespace, ErrNamespaceNotLegacy when the namespace is already new, ErrForbidden when sc was
	// never constructed, and the store's error otherwise, with the upgrade rolled back when the
	// failure comes inside the transaction.
	UpgradeNamespace(ctx context.Context, sc scope.Scope, tenantID string) error
}

func (s *service) CreateNamespace(ctx context.Context, req *requests.NamespaceCreate) (*models.Namespace, error) {
	var ns *models.Namespace
	if err := s.store.WithTransaction(ctx, func(ctx context.Context) error {
		var err error
		ns, err = s.createNamespaceInTransaction(ctx, req)

		return err
	}); err != nil {
		return nil, err
	}

	return ns, nil
}

func (s *service) createNamespaceInTransaction(ctx context.Context, req *requests.NamespaceCreate) (*models.Namespace, error) {
	user, err := s.store.UserResolve(ctx, store.UserIDResolver, req.UserID)
	if err != nil || user == nil {
		return nil, NewErrUserNotFound(req.UserID, err)
	}

	if user.MaxNamespaces == 0 {
		return nil, NewErrNamespaceCreationIsForbidden(user.MaxNamespaces, nil)
	} else if user.MaxNamespaces > 0 {
		if err := s.store.NamespaceLockOwner(ctx, user.ID); err != nil {
			return nil, err
		}

		info, err := s.store.UserGetInfo(ctx, user.ID)
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
			SessionRecord:            true,
			ConnectionAnnouncement:   envs.AnnouncementFor(s.edition),
			RequireVerifiedDeviceKey: s.namespaceRequireVerifiedDeviceKey,
		},
		Type: models.NewDefaultType(),
	}

	if models.IsTypeTeam(req.Type) {
		ns.Type = models.TypeTeam
	} else if models.IsTypePersonal(req.Type) {
		ns.Type = models.TypePersonal
	}

	if s.edition.IsCloud() {
		ns.MaxDevices = 3
	} else {
		ns.MaxDevices = -1
	}

	tenantID, err := s.store.NamespaceCreate(ctx, ns)
	if err != nil {
		if errors.Is(err, store.ErrDuplicate) {
			return nil, NewErrNamespaceDuplicated(err)
		}

		if errors.Is(err, store.ErrNamespaceSingle) {
			return nil, NewErrNamespaceSingle(err)
		}

		return nil, NewErrNamespaceCreateStore(err)
	}

	ns.TenantID = tenantID

	return ns, nil
}

func (s *service) ListNamespaces(ctx context.Context, req *requests.NamespaceList) ([]models.Namespace, int64, error) {
	for p := range req.Filters.Properties() {
		if col, found := namespaceFilterColumns[p.Name]; found {
			p.Name = col
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

func (s *service) GetNamespace(ctx context.Context, sc scope.Scope, tenantID string) (*models.Namespace, error) {
	return s.resolveWithin(ctx, sc, tenantID, anyResolveError)
}

func (s *service) ListNamespaceMembers(ctx context.Context, sc scope.Scope, req *requests.MemberList) ([]models.MemberView, int64, error) {
	namespace, err := s.resolveWithin(ctx, sc, req.TenantID, anyResolveError)
	if err != nil {
		return nil, 0, err
	}

	bounded := scope.MustBounded(namespace.TenantID)

	opts := []store.QueryOption{s.store.Options().Paginate(&req.Paginator)}

	return s.store.NamespaceGetMembers(ctx, bounded, opts...)
}

func (s *service) DeleteNamespace(ctx context.Context, sc scope.Scope, tenantID string) error {
	n, err := s.resolveWithin(ctx, sc, tenantID, anyResolveError)
	if err != nil {
		return err
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

func (s *service) EditNamespace(ctx context.Context, sc scope.Scope, req *requests.NamespaceEdit) (*models.Namespace, error) {
	namespace, err := s.resolveWithin(ctx, sc, req.Tenant, anyResolveError)
	if err != nil {
		return nil, err
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

	if req.Settings.RequireVerifiedDeviceKey != nil {
		namespace.Settings.RequireVerifiedDeviceKey = *req.Settings.RequireVerifiedDeviceKey
	}

	if err := s.store.NamespaceUpdate(ctx, namespace); err != nil {
		if errors.Is(err, store.ErrDuplicate) {
			return nil, NewErrNamespaceDuplicated(err)
		}

		return nil, err
	}

	return s.store.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, req.Tenant)
}

func (s *service) EditSessionRecordStatus(ctx context.Context, sc scope.Scope, sessionRecord bool, tenantID string) error {
	n, err := s.resolveWithin(ctx, sc, tenantID, missingDocument)
	if err != nil {
		return err
	}

	n.Settings.SessionRecord = sessionRecord
	if err := s.store.NamespaceUpdate(ctx, n); err != nil {
		return err
	}

	return nil
}

func (s *service) UpgradeNamespace(ctx context.Context, sc scope.Scope, tenantID string) error {
	n, err := s.resolveWithin(ctx, sc, tenantID, missingDocument)
	if err != nil {
		return err
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
