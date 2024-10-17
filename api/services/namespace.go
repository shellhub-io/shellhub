package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/mongo"
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

	// EditNamespace updates a namespace for the specified requests.NamespaceEdit#Tenant.
	// It returns the namespace with the updated fields and an error, if any.
	EditNamespace(ctx context.Context, req *requests.NamespaceEdit) (*models.Namespace, error)

	// AddNamespaceMember adds a member to a namespace.
	//
	// In cloud environments, the member is assigned a [MemberStatusPending] status until they accept the invite via
	// an invitation email. If the target user does not exist, the email will redirect them to the registration page,
	// and the invite can be accepted after finishing. In community and enterprise environments, the status is set to
	// [MemberStatusAccepted] without sending an email.
	//
	// The role assigned to the new member must not grant more authority than the user adding them (e.g.,
	// an administrator cannot add a member with a higher role such as an owner). Owners cannot be created.
	//
	// It returns the namespace and an error, if any.
	AddNamespaceMember(ctx context.Context, req *requests.NamespaceAddMember) (*models.Namespace, error)

	// UpdateNamespaceMember updates a member with the specified ID in the specified namespace. The member's role cannot
	// have more authority than the user who is updating the member; owners cannot be created. It returns an error, if any.
	UpdateNamespaceMember(ctx context.Context, req *requests.NamespaceUpdateMember) error
	// RemoveNamespaceMember removes a member with the specified ID in the specified namespace. The member's role cannot
	// have more authority than the user who is removing the member; owners cannot be removed. It returns the namespace
	// and an error, if any.
	RemoveNamespaceMember(ctx context.Context, req *requests.NamespaceRemoveMember) (*models.Namespace, error)

	EditSessionRecordStatus(ctx context.Context, sessionRecord bool, tenantID string) error
	GetSessionRecord(ctx context.Context, tenantID string) (bool, error)
}

// CreateNamespace creates a new namespace.
func (s *service) CreateNamespace(ctx context.Context, req *requests.NamespaceCreate) (*models.Namespace, error) {
	user, _, err := s.store.UserGetByID(ctx, req.UserID, false)
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

	if dup, err := s.store.NamespaceGetByName(ctx, strings.ToLower(req.Name)); dup != nil || (err != nil && err != store.ErrNoDocuments) {
		return nil, NewErrNamespaceDuplicated(err)
	}

	ns := &models.Namespace{
		Name:  strings.ToLower(req.Name),
		Owner: user.ID,
		Members: []models.Member{
			{
				ID:      user.ID,
				Role:    authorizer.RoleOwner,
				Status:  models.MemberStatusAccepted,
				AddedAt: clock.Now(),
			},
		},
		Settings: &models.NamespaceSettings{
			SessionRecord:          true,
			ConnectionAnnouncement: "",
		},
		TenantID: req.TenantID,
	}

	if envs.IsCommunity() {
		ns.Settings.ConnectionAnnouncement = models.DefaultAnnouncementMessage
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
	namespaces, count, err := s.store.NamespaceList(ctx, req.Paginator, req.Filters, false, s.store.Options().CountAcceptedDevices(), s.store.Options().EnrichMembersData())
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
	namespace, err := s.store.NamespaceGet(ctx, tenantID, s.store.Options().CountAcceptedDevices(), s.store.Options().EnrichMembersData())
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
	ns, err := s.store.NamespaceGet(ctx, tenantID, s.store.Options().CountAcceptedDevices())
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

	return s.store.NamespaceDelete(ctx, tenantID)
}

func (s *service) EditNamespace(ctx context.Context, req *requests.NamespaceEdit) (*models.Namespace, error) {
	changes := &models.NamespaceChanges{
		Name:                   strings.ToLower(req.Name),
		SessionRecord:          req.Settings.SessionRecord,
		ConnectionAnnouncement: req.Settings.ConnectionAnnouncement,
	}

	if err := s.store.NamespaceEdit(ctx, req.Tenant, changes); err != nil {
		switch {
		case errors.Is(err, store.ErrNoDocuments):
			return nil, NewErrNamespaceNotFound(req.Tenant, err)
		default:
			return nil, err
		}
	}

	return s.store.NamespaceGet(ctx, req.Tenant, s.store.Options().CountAcceptedDevices(), s.store.Options().EnrichMembersData())
}

func (s *service) AddNamespaceMember(ctx context.Context, req *requests.NamespaceAddMember) (*models.Namespace, error) {
	namespace, err := s.store.NamespaceGet(ctx, req.TenantID)
	if err != nil || namespace == nil {
		return nil, NewErrNamespaceNotFound(req.TenantID, err)
	}

	user, _, err := s.store.UserGetByID(ctx, req.UserID, false)
	if err != nil || user == nil {
		return nil, NewErrUserNotFound(req.UserID, err)
	}

	// checks if the active member is in the namespace. user is the active member.
	active, ok := namespace.FindMember(user.ID)
	if !ok {
		return nil, NewErrNamespaceMemberNotFound(user.ID, err)
	}

	if !active.Role.HasAuthority(req.MemberRole) {
		return nil, NewErrRoleInvalid()
	}

	// In cloud instances, if the target user does not exist, we need to create a new user
	// with the specified email. We use the inserted ID to identify the user once they complete
	// the registration and accepts the invitation.
	passiveUser, err := s.store.UserGetByEmail(ctx, strings.ToLower(req.MemberEmail))
	if err != nil {
		if !envs.IsCloud() || !errors.Is(err, store.ErrNoDocuments) {
			return nil, NewErrUserNotFound(req.MemberEmail, err)
		}

		passiveUser = &models.User{}
		passiveUser.ID, err = s.store.UserCreateInvited(ctx, strings.ToLower(req.MemberEmail))
		if err != nil {
			return nil, err
		}
	}

	// In cloud instances, if a member exists and their status is pending and the expiration date is reached,
	// we resend the invite instead of adding the member.
	// In community and enterprise instances, a "duplicate" error is always returned,
	// since the member will never be in a pending status.
	// Otherwise, add the member "from scratch"
	if m, ok := namespace.FindMember(passiveUser.ID); ok {
		now := clock.Now()

		if !envs.IsCloud() || !(m.Status == models.MemberStatusPending && m.ExpiresAt.Before(now)) {
			return nil, NewErrNamespaceMemberDuplicated(passiveUser.ID, nil)
		}

		if err := s.store.WithTransaction(ctx, s.resendMemberInvite(m.ID, req)); err != nil {
			return nil, err
		}
	} else {
		if err := s.store.WithTransaction(ctx, s.addMember(passiveUser.ID, req)); err != nil {
			return nil, err
		}
	}

	return s.store.NamespaceGet(ctx, req.TenantID, s.store.Options().CountAcceptedDevices(), s.store.Options().EnrichMembersData())
}

// addMember returns a transaction callback that adds a member and sends an invite if the instance is cloud.
func (s *service) addMember(memberID string, req *requests.NamespaceAddMember) store.TransactionCb {
	return func(ctx context.Context) error {
		member := &models.Member{
			ID:      memberID,
			AddedAt: clock.Now(),
			Role:    req.MemberRole,
		}

		// In cloud instances, the member must accept the invite before enter in the namespace.
		if envs.IsCloud() {
			member.Status = models.MemberStatusPending
			member.ExpiresAt = member.AddedAt.Add(7 * (24 * time.Hour))
		} else {
			member.Status = models.MemberStatusAccepted
			member.ExpiresAt = time.Time{}
		}

		if err := s.store.NamespaceAddMember(ctx, req.TenantID, member); err != nil {
			return err
		}

		if envs.IsCloud() {
			if err := s.client.InviteMember(ctx, req.TenantID, member.ID, req.FowardedHost); err != nil {
				return err
			}
		}

		return nil
	}
}

// resendMemberInvite returns a transaction callback that resends an invitation to the member with the
// specified ID.
func (s *service) resendMemberInvite(memberID string, req *requests.NamespaceAddMember) store.TransactionCb {
	return func(ctx context.Context) error {
		expiresAt := clock.Now().Add(7 * (24 * time.Hour))
		changes := &models.MemberChanges{ExpiresAt: &expiresAt, Role: req.MemberRole}

		if err := s.store.NamespaceUpdateMember(ctx, req.TenantID, memberID, changes); err != nil {
			return err
		}

		return s.client.InviteMember(ctx, req.TenantID, memberID, req.FowardedHost)
	}
}

func (s *service) UpdateNamespaceMember(ctx context.Context, req *requests.NamespaceUpdateMember) error {
	namespace, err := s.store.NamespaceGet(ctx, req.TenantID)
	if err != nil {
		return NewErrNamespaceNotFound(req.TenantID, err)
	}

	user, _, err := s.store.UserGetByID(ctx, req.UserID, false)
	if err != nil {
		return NewErrUserNotFound(req.UserID, err)
	}

	active, ok := namespace.FindMember(user.ID)
	if !ok {
		return NewErrNamespaceMemberNotFound(user.ID, err)
	}

	if _, ok := namespace.FindMember(req.MemberID); !ok {
		return NewErrNamespaceMemberNotFound(req.MemberID, err)
	}

	changes := &models.MemberChanges{Role: req.MemberRole}

	if changes.Role != authorizer.RoleInvalid {
		if !active.Role.HasAuthority(req.MemberRole) {
			return NewErrRoleInvalid()
		}
	}

	if err := s.store.NamespaceUpdateMember(ctx, req.TenantID, req.MemberID, changes); err != nil {
		return err
	}

	s.AuthUncacheToken(ctx, namespace.TenantID, req.MemberID) // nolint: errcheck

	return nil
}

func (s *service) RemoveNamespaceMember(ctx context.Context, req *requests.NamespaceRemoveMember) (*models.Namespace, error) {
	namespace, err := s.store.NamespaceGet(ctx, req.TenantID)
	if err != nil {
		return nil, NewErrNamespaceNotFound(req.TenantID, err)
	}

	user, _, err := s.store.UserGetByID(ctx, req.UserID, false)
	if err != nil {
		return nil, NewErrUserNotFound(req.UserID, err)
	}

	active, ok := namespace.FindMember(user.ID)
	if !ok {
		return nil, NewErrNamespaceMemberNotFound(user.ID, err)
	}

	passive, ok := namespace.FindMember(req.MemberID)
	if !ok {
		return nil, NewErrNamespaceMemberNotFound(req.MemberID, err)
	}

	if !active.Role.HasAuthority(passive.Role) {
		return nil, NewErrRoleInvalid()
	}

	if err := s.store.NamespaceRemoveMember(ctx, req.TenantID, req.MemberID); err != nil {
		switch {
		case errors.Is(err, store.ErrNoDocuments):
			return nil, NewErrNamespaceNotFound(req.TenantID, err)
		case errors.Is(err, mongo.ErrUserNotFound):
			return nil, NewErrNamespaceMemberNotFound(req.MemberID, err)
		default:
			return nil, err
		}
	}

	s.AuthUncacheToken(ctx, namespace.TenantID, req.MemberID) // nolint: errcheck

	return s.store.NamespaceGet(ctx, req.TenantID, s.store.Options().CountAcceptedDevices(), s.store.Options().EnrichMembersData())
}

// EditSessionRecordStatus defines if the sessions will be recorded.
//
// It receives a context, used to "control" the request flow, a boolean to define if the sessions will be recorded and
// the tenant ID from models.Namespace.
func (s *service) EditSessionRecordStatus(ctx context.Context, sessionRecord bool, tenantID string) error {
	return s.store.NamespaceSetSessionRecord(ctx, sessionRecord, tenantID)
}

// GetSessionRecord gets the session record data.
//
// It receives a context, used to "control" the request flow, the tenant ID from models.Namespace.
//
// GetSessionRecord returns a boolean indicating the session record status and an error. When error is not nil,
// the boolean is false.
func (s *service) GetSessionRecord(ctx context.Context, tenantID string) (bool, error) {
	if _, err := s.store.NamespaceGet(ctx, tenantID); err != nil {
		return false, NewErrNamespaceNotFound(tenantID, err)
	}

	return s.store.NamespaceGetSessionRecord(ctx, tenantID)
}
