package services

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strings"

	"github.com/shellhub-io/shellhub/api/pkg/guard"
	"github.com/shellhub-io/shellhub/api/store"
	req "github.com/shellhub-io/shellhub/pkg/api/internalclient"
	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/authorizer"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/models"
	hp "github.com/shellhub-io/shellhub/pkg/requests"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	"github.com/shellhub-io/shellhub/pkg/validator"
)

type NamespaceService interface {
	ListNamespaces(ctx context.Context, pagination paginator.Query, filterB64 string, export bool) ([]models.Namespace, int, error)
	CreateNamespace(ctx context.Context, namespace *models.Namespace, userID string) (*models.Namespace, error)
	GetNamespace(ctx context.Context, tenantID string) (*models.Namespace, error)
	DeleteNamespace(ctx context.Context, tenantID string) error
	EditNamespace(ctx context.Context, tenantID, name string) (*models.Namespace, error)
	AddNamespaceUser(ctx context.Context, memberUsername, memberRole, tenantID, userID string) (*models.Namespace, error)
	RemoveNamespaceUser(ctx context.Context, tenantID, memberID, userID string) (*models.Namespace, error)
	EditNamespaceUser(ctx context.Context, tenantID, userID, memberID, memberNewRole string) error
	ListMembers(ctx context.Context, tenantID string) ([]models.Member, error)
	EditSessionRecordStatus(ctx context.Context, sessionRecord bool, tenantID string) error
	GetSessionRecord(ctx context.Context, tenantID string) (bool, error)
	HandleReportDelete(ns *models.Namespace) error
}

func (s *service) HandleReportDelete(ns *models.Namespace) error {
	if !hp.HasBillingInstance(ns) {
		return nil
	}

	status, err := s.client.(req.Client).ReportDelete(ns)
	if err != nil {
		return err
	}

	return HandleStatusResponse(status)
}

// ListNamespaces lists all namespace from a user.
//
// pagination is a paginator.Query defines how many namespaces will be returned, filterB64 is a JSON object encoded
// in B64 used to make filter's operation and export is.
func (s *service) ListNamespaces(ctx context.Context, pagination paginator.Query, filterB64 string, export bool) ([]models.Namespace, int, error) {
	raw, err := base64.StdEncoding.DecodeString(filterB64)
	if err != nil {
		return nil, 0, err
	}

	var filter []models.Filter

	if err := json.Unmarshal(raw, &filter); len(raw) > 0 && err != nil {
		return nil, 0, err
	}

	namespaces, count, err := s.store.NamespaceList(ctx, pagination, filter, export)
	if err != nil {
		return nil, 0, err
	}

	for count, namespace := range namespaces {
		members, err := s.ListMembers(ctx, namespace.TenantID)
		if err != nil {
			return nil, 0, err
		}

		namespaces[count].Members = []models.Member{}
		namespaces[count].Members = append(namespaces[count].Members, members...)
	}

	return namespaces, count, nil
}

// CreateNamespace creates a new namespace.
//
// namespace is the model what you want to add and userID is the models.User's ID who will get the namespace.
func (s *service) CreateNamespace(ctx context.Context, namespace *models.Namespace, userID string) (*models.Namespace, error) {
	user, _, err := s.store.UserGetByID(ctx, userID, false)
	if user == nil {
		return nil, ErrForbidden
	}

	if err != nil {
		return nil, err
	}

	ns := &models.Namespace{
		Name:  strings.ToLower(namespace.Name),
		Owner: user.ID,
		Members: []models.Member{
			{
				ID:   user.ID,
				Role: authorizer.MemberRoleOwner,
			},
		},
		Settings: &models.NamespaceSettings{SessionRecord: true},
		TenantID: namespace.TenantID,
	}

	if _, err := validator.ValidateStruct(ns); err != nil {
		return nil, ErrInvalidFormat
	}

	if namespace.TenantID == "" {
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

	otherNamespace, err := s.store.NamespaceGetByName(ctx, ns.Name)
	if err != nil && err != store.ErrNoDocuments {
		return nil, err
	}

	if otherNamespace != nil {
		return nil, ErrConflictName
	}

	if _, err := s.store.NamespaceCreate(ctx, ns); err != nil {
		return nil, err
	}

	return ns, nil
}

// GetNamespace gets a namespace.
//
// tenantID is the models.Namespace's tenant what you want to get.
func (s *service) GetNamespace(ctx context.Context, tenantID string) (*models.Namespace, error) {
	namespaces, err := s.store.NamespaceGet(ctx, tenantID)
	if err != nil || namespaces == nil {
		return nil, ErrNamespaceNotFound
	}

	members := []models.Member{}
	for _, member := range namespaces.Members {
		user, _, err := s.store.UserGetByID(ctx, member.ID, false)
		if err != nil {
			if err == store.ErrNoDocuments {
				return nil, ErrUserNotFound
			}

			return nil, err
		}

		member := models.Member{ID: user.ID, Username: user.Username, Role: member.Role}
		members = append(members, member)
	}

	namespaces.Members = []models.Member{}
	namespaces.Members = append(namespaces.Members, members...)

	return namespaces, nil
}

// DeleteNamespace deletes a namespace.
//
// tenantID is the models.Namespace's tenant what you want to delete.
func (s *service) DeleteNamespace(ctx context.Context, tenantID string) error {
	ns, err := s.store.NamespaceGet(ctx, tenantID)
	if err != nil {
		return err
	}
	if err := s.HandleReportDelete(ns); err != nil {
		return err
	}

	return s.store.NamespaceDelete(ctx, tenantID)
}

// ListMembers lists all members from a namespace.
//
// tenantID is the models.Namespace's tenant what you want the members.
func (s *service) ListMembers(ctx context.Context, tenantID string) ([]models.Member, error) {
	ns, err := s.store.NamespaceGet(ctx, tenantID)
	if err != nil {
		if err == store.ErrNoDocuments {
			return nil, ErrNamespaceNotFound
		}

		return nil, err
	}

	members := []models.Member{}
	for _, member := range ns.Members {
		user, _, err := s.store.UserGetByID(ctx, member.ID, false)
		if err != nil {
			if err == store.ErrNoDocuments {
				return nil, ErrUserNotFound
			}

			return nil, err
		}

		member := models.Member{ID: user.ID, Username: user.Username, Role: member.Role}
		members = append(members, member)
	}

	return members, nil
}

// EditNamespace changes the namespace's name.
//
// tenantID is the models.Namespace's tenant what will be the name changed and name is new name.
func (s *service) EditNamespace(ctx context.Context, tenantID, name string) (*models.Namespace, error) {
	ns, err := s.store.NamespaceGet(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	lowerName := strings.ToLower(name)
	if _, err := validator.ValidateStruct(&models.Namespace{
		Name: lowerName,
	}); err != nil {
		return nil, ErrInvalidFormat
	}

	if ns.Name == lowerName {
		return nil, ErrBadRequest
	}

	return s.store.NamespaceRename(ctx, ns.TenantID, lowerName)
}

// AddNamespaceUser adds an user to a namespace.
//
// username is the models.User's name from the member what you want to add, role is member's role, tenantID is the
// models.Namespace's tenant what will receive the member and userID is the models.User's ID who is adding this member.
func (s *service) AddNamespaceUser(ctx context.Context, memberUsername, memberRole, tenantID, userID string) (*models.Namespace, error) {
	if _, err := validator.ValidateStruct(models.Member{Username: memberUsername, Role: memberRole}); err != nil {
		return nil, ErrInvalidFormat
	}

	namespace, err := s.store.NamespaceGet(ctx, tenantID)
	if err != nil {
		if err == store.ErrNoDocuments {
			return nil, ErrNamespaceNotFound
		}

		return nil, err
	}

	memberActive, _, err := s.store.UserGetByID(ctx, userID, false)
	if err != nil {
		if err == store.ErrNoDocuments {
			return nil, ErrUserNotFound
		}

		return nil, err
	}

	// Checks if the active member is in the namespace.
	memberActiveFound, ok := guard.CheckMember(namespace, memberActive.ID)
	if !ok {
		return nil, ErrBadRequest
	}

	memberPassive, err := s.store.UserGetByUsername(ctx, memberUsername)
	if err != nil {
		if err == store.ErrNoDocuments {
			return nil, ErrUserNotFound
		}

		return nil, err
	}

	// Checks if the passive member is in the namespace.
	_, ok = guard.CheckMember(namespace, memberPassive.ID)
	if ok {
		return nil, ErrNamespaceDuplicatedMember
	}

	if !guard.CheckRole(memberActiveFound.Role, memberRole) {
		return nil, guard.ErrForbidden
	}

	return s.store.NamespaceAddMember(ctx, tenantID, memberPassive.ID, memberRole)
}

// RemoveNamespaceUser removes a user from a namespace.
//
// tenantID is the models.Namespace's tenant what will remove the member, mid is the member who will be removed and userID is
// the models.User's ID from who is removing the member.
func (s *service) RemoveNamespaceUser(ctx context.Context, tenantID, memberID, userID string) (*models.Namespace, error) {
	namespace, err := s.store.NamespaceGet(ctx, tenantID)
	if err != nil {
		if err == store.ErrNoDocuments {
			return nil, ErrNamespaceNotFound
		}

		return nil, err
	}

	memberActive, _, err := s.store.UserGetByID(ctx, userID, false)
	if err != nil {
		if err == store.ErrNoDocuments {
			return nil, ErrUserNotFound
		}

		return nil, err
	}

	memberPassive, _, err := s.store.UserGetByID(ctx, memberID, false)
	if err != nil {
		if err == store.ErrNoDocuments {
			return nil, ErrUserNotFound
		}

		return nil, err
	}

	memberActiveFound, okActive := guard.CheckMember(namespace, memberActive.ID)
	memberPassiveFound, okPassive := guard.CheckMember(namespace, memberPassive.ID)
	if !okActive || !okPassive {
		return nil, ErrNamespaceMemberNotFound
	}

	if !guard.CheckRole(memberActiveFound.Role, memberPassiveFound.Role) {
		return nil, guard.ErrForbidden
	}

	return s.store.NamespaceRemoveMember(ctx, tenantID, memberPassive.ID)
}

// EditNamespaceUser changes user's role.
//
// tenantID is the models.Namespace's tenant from what namespace you want to edit, userID is the models.User's ID from who is
// acting to change the role, mid is the models.Member's ID who will be the role changed and role is the new role to
// member.
func (s *service) EditNamespaceUser(ctx context.Context, tenantID, userID, memberID, memberNewRole string) error {
	namespace, err := s.store.NamespaceGet(ctx, tenantID)
	if err != nil {
		if err == store.ErrNoDocuments {
			return ErrNamespaceNotFound
		}

		return err
	}

	memberActive, _, err := s.store.UserGetByID(ctx, userID, false)
	if err != nil {
		if err == store.ErrNoDocuments {
			return ErrUserNotFound
		}

		return err
	}

	memberPassive, _, err := s.store.UserGetByID(ctx, memberID, false)
	if err != nil {
		if err == store.ErrNoDocuments {
			return ErrUserNotFound
		}

		return err
	}

	memberActiveFound, okActive := guard.CheckMember(namespace, memberActive.ID)
	memberPassiveFound, okPassive := guard.CheckMember(namespace, memberPassive.ID)
	if !okActive || !okPassive {
		return ErrNamespaceMemberNotFound
	}

	// Blocks if the active member's role is equal to the passive one.
	if memberPassiveFound.Role == memberActiveFound.Role {
		return guard.ErrForbidden
	}

	if !guard.CheckRole(memberActiveFound.Role, memberNewRole) {
		return guard.ErrForbidden
	}

	return s.store.NamespaceEditMember(ctx, tenantID, memberPassive.ID, memberNewRole)
}

// EditSessionRecordStatus defines if the session will be recorded.
//
// record is the state what define if there will session record in a namespace and tenantID is the models.Namespace's tenant
// from what namespace you want to record.
func (s *service) EditSessionRecordStatus(ctx context.Context, sessionRecord bool, tenantID string) error {
	return s.store.NamespaceSetSessionRecord(ctx, sessionRecord, tenantID)
}

// GetSessionRecord gets the session record state.
//
// tenantID is the models.Namespace's tenant to get the session record status.
func (s *service) GetSessionRecord(ctx context.Context, tenantID string) (bool, error) {
	if _, err := s.store.NamespaceGet(ctx, tenantID); err != nil {
		if err == store.ErrNoDocuments {
			return false, ErrNamespaceNotFound
		}

		return false, err
	}

	return s.store.NamespaceGetSessionRecord(ctx, tenantID)
}
