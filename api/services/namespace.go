package services

import (
	"context"
	"errors"
	"strings"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
)

type NamespaceService interface {
	ListNamespaces(ctx context.Context, req *requests.NamespaceList) ([]models.Namespace, int, error)
	CreateNamespace(ctx context.Context, namespace *requests.NamespaceCreate) (*models.Namespace, error)
	GetNamespace(ctx context.Context, tenantID string) (*models.Namespace, error)
	DeleteNamespace(ctx context.Context, tenantID string) error
	EditSessionRecordStatus(ctx context.Context, sessionRecord bool, tenantID string) error
	GetSessionRecord(ctx context.Context, tenantID string) (bool, error)
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

	if _, err := s.store.NamespaceCreate(ctx, ns); err != nil {
		return nil, NewErrNamespaceCreateStore(err)
	}

	return ns, nil
}

func (s *service) ListNamespaces(ctx context.Context, req *requests.NamespaceList) ([]models.Namespace, int, error) {
	opts := []store.QueryOption{s.store.Options().Match(&req.Filters), s.store.Options().Paginate(&req.Paginator)}
	// NOTE: WithMember filters namespaces where the user is a member. In Mongo this is a no-op
	// because the filtering is done implicitly via gateway.IDFromContext inside the store.
	// In PostgreSQL, WithMember applies the filter explicitly.
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
		if err := s.BillingReport(ctx, tenantID, ReportNamespaceDelete); err != nil {
			return NewErrBillingReportNamespaceDelete(err)
		}
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

// GetSessionRecord gets the session record data.
//
// It receives a context, used to "control" the request flow, the tenant ID from models.Namespace.
//
// GetSessionRecord returns a boolean indicating the session record status and an error. When error is not nil,
// the boolean is false.
func (s *service) GetSessionRecord(ctx context.Context, tenantID string) (bool, error) {
	n, err := s.store.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, tenantID)
	if err != nil {
		return false, NewErrNamespaceNotFound(tenantID, err)
	}

	return n.Settings.SessionRecord, nil
}
