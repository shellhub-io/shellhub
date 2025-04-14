package services

import (
	"context"
	"strings"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
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
	user, err := s.store.UserGet(ctx, store.UserIdentID, req.UserID)
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

	req.Name = strings.ToLower(req.Name)

	if _, has, err := s.store.NamespaceConflicts(ctx, &models.NamespaceConflicts{Name: req.Name}); err != nil || has {
		return nil, NewErrNamespaceDuplicated(err)
	}

	if req.TenantID == "" {
		req.TenantID = uuid.Generate()
	}

	ns := &models.Namespace{
		TenantID: req.TenantID,
		Type:     models.TypeTeam,
		Name:     strings.ToLower(req.Name),
		Settings: &models.NamespaceSettings{},
		Members: []models.Member{
			{
				ID:     user.ID,
				Status: models.MemberStatusAccepted,
				Role:   authorizer.RoleOwner,
			},
		},
	}

	// Set limits according to ShellHub instance type
	if envs.IsCloud() {
		// cloud free plan is limited only by the max of devices
		ns.MaxDevices = 3
		ns.Settings.ConnectionAnnouncement = ""
		ns.Settings.SessionRecord = true
	} else {
		// we don't set limits on enterprise and community instances
		ns.MaxDevices = -1
		ns.Settings.ConnectionAnnouncement = models.DefaultAnnouncementMessage
		ns.Settings.SessionRecord = envs.IsEnterprise()
	}

	// TODO: transactions
	if _, err := s.store.NamespaceCreate(ctx, ns); err != nil {
		return nil, NewErrNamespaceCreateStore(err)
	}

	if err := s.store.NamespaceCreateMemberships(ctx, ns.TenantID, ns.Members...); err != nil {
		return nil, err
	}

	return ns, nil
}

func (s *service) ListNamespaces(ctx context.Context, req *requests.NamespaceList) ([]models.Namespace, int, error) {
	namespaces, count, err := s.store.NamespaceList(
		ctx,
		s.store.Options().WithMember(req.UserID),
		s.store.Options().Paginate(req.Paginator),
		s.store.Options().Order(req.Sorter),
		s.store.Options().Filter(req.Filters),
	)
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
	namespace, err := s.store.NamespaceGet(ctx, store.NamespaceIdentTenantID, tenantID)
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
	ns, err := s.store.NamespaceGet(ctx, store.NamespaceIdentTenantID, tenantID)
	if err != nil {
		return NewErrNamespaceNotFound(tenantID, err)
	}

	ableToReportDeleteNamespace := func(ns *models.Namespace) bool {
		return !ns.Billing.IsNil() && ns.Billing.HasCutomer() && ns.Billing.HasSubscription()
	}

	if envs.IsCloud() && envs.HasBilling() && ableToReportDeleteNamespace(ns) {
		if err := s.BillingReport(s.client, tenantID, ReportNamespaceDelete); err != nil {
			return NewErrBillingReportNamespaceDelete(err)
		}
	}

	return s.store.NamespaceDelete(ctx, ns)
}

func (s *service) EditNamespace(ctx context.Context, req *requests.NamespaceEdit) (*models.Namespace, error) {
	ns, err := s.store.NamespaceGet(ctx, store.NamespaceIdentTenantID, req.TenantID)
	if err != nil {
		return nil, NewErrNamespaceNotFound(req.TenantID, err)
	}

	if req.Name != "" {
		ns.Name = strings.ToLower(req.Name)
	}

	if req.Settings.SessionRecord != nil {
		ns.Settings.SessionRecord = *req.Settings.SessionRecord
	}

	if req.Settings.ConnectionAnnouncement != nil {
		ns.Settings.ConnectionAnnouncement = *req.Settings.ConnectionAnnouncement
	}

	if err := s.store.NamespaceSave(ctx, ns); err != nil {
		return nil, err
	}

	return ns, nil
}

// EditSessionRecordStatus defines if the sessions will be recorded.
//
// It receives a context, used to "control" the request flow, a boolean to define if the sessions will be recorded and
// the tenant ID from models.Namespace.
func (s *service) EditSessionRecordStatus(ctx context.Context, sessionRecord bool, tenantID string) error {
	// return s.store.NamespaceSetSessionRecord(ctx, sessionRecord, tenantID)
	return nil
}

// GetSessionRecord gets the session record data.
//
// It receives a context, used to "control" the request flow, the tenant ID from models.Namespace.
//
// GetSessionRecord returns a boolean indicating the session record status and an error. When error is not nil,
// the boolean is false.
func (s *service) GetSessionRecord(ctx context.Context, tenantID string) (bool, error) {
	// if _, err := s.store.NamespaceGet(ctx, store.NamespaceIdentID, tenantID); err != nil {
	// 	return false, NewErrNamespaceNotFound(tenantID, err)
	// }
	//
	// return s.store.NamespaceGetSessionRecord(ctx, tenantID)
	return false, nil
}
