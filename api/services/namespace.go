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
	AddNamespaceUser(ctx context.Context, memberUsername, memberType, tenantID, userID string) (*models.Namespace, error)
	RemoveNamespaceUser(ctx context.Context, tenantID, memberID, userID string) (*models.Namespace, error)
	EditNamespaceUser(ctx context.Context, tenantID, userID, memberID, memberNewType string) error
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

	return hp.HandleStatusResponse(status)
}

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
				Type: authorizer.MemberTypeOwner,
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

func (s *service) GetNamespace(ctx context.Context, tenantID string) (*models.Namespace, error) {
	namespaces, err := s.store.NamespaceGet(ctx, tenantID)
	if err != nil {
		return nil, err
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

		member := models.Member{ID: user.ID, Username: user.Username, Type: member.Type}
		members = append(members, member)
	}

	namespaces.Members = []models.Member{}
	namespaces.Members = append(namespaces.Members, members...)

	return namespaces, nil
}

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

		member := models.Member{ID: user.ID, Username: user.Username, Type: member.Type}
		members = append(members, member)
	}

	return members, nil
}

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

func (s *service) AddNamespaceUser(ctx context.Context, memberUsername, memberType, tenantID, userID string) (*models.Namespace, error) {
	findMemberNamespace := func(member *models.User, namespace *models.Namespace) (*models.Member, bool) {
		var memberFound models.Member
		for _, memberSearch := range namespace.Members {
			if memberSearch.ID == member.ID {
				memberFound = memberSearch
			}
		}
		if memberFound.ID == "" || memberFound.Type == "" {
			return nil, false
		}

		return &memberFound, true
	}

	if _, err := validator.ValidateStruct(models.Member{Username: memberUsername, Type: memberType}); err != nil {
		return nil, ErrInvalidFormat
	}

	if !guard.EvaluateSubject(ctx, s.store, tenantID, userID, memberType) {
		return nil, ErrForbidden
	}

	member, err := s.store.UserGetByUsername(ctx, memberUsername)
	if err != nil {
		if err == store.ErrNoDocuments {
			return nil, ErrUserNotFound
		}

		return nil, err
	}

	namespace, err := s.store.NamespaceGet(ctx, tenantID)
	if err != nil {
		if err == store.ErrNoDocuments {
			return nil, ErrNamespaceNotFound
		}

		return nil, err
	}

	_, ok := findMemberNamespace(member, namespace)
	if ok {
		return nil, ErrNamespaceDuplicatedMember
	}

	return s.store.NamespaceAddMember(ctx, tenantID, member.ID, memberType)
}

func (s *service) RemoveNamespaceUser(ctx context.Context, tenantID, memberID, userID string) (*models.Namespace, error) {
	findMemberNamespace := func(member *models.User, namespace *models.Namespace) (*models.Member, bool) {
		var memberFound models.Member
		for _, memberSearch := range namespace.Members {
			if memberSearch.ID == member.ID {
				memberFound = memberSearch
			}
		}
		if memberFound.ID == "" || memberFound.Type == "" {
			return nil, false
		}

		return &memberFound, true
	}

	memberPassive, _, err := s.store.UserGetByID(ctx, memberID, false)
	if err != nil {
		if err == store.ErrNoDocuments {
			return nil, ErrUserNotFound
		}

		return nil, err
	}

	namespace, err := s.store.NamespaceGet(ctx, tenantID)
	if err != nil {
		if err == store.ErrNoDocuments {
			return nil, ErrNamespaceNotFound
		}

		return nil, err
	}

	memberFound, ok := findMemberNamespace(memberPassive, namespace)
	if !ok {
		return nil, ErrNamespaceMemberNotFound
	}

	if !guard.EvaluateSubject(ctx, s.store, tenantID, userID, memberFound.Type) {
		return nil, ErrForbidden
	}

	return s.store.NamespaceRemoveMember(ctx, tenantID, memberPassive.ID)
}

func (s *service) EditNamespaceUser(ctx context.Context, tenantID, userID, memberID, memberNewType string) error {
	findMemberNamespace := func(member *models.User, namespace *models.Namespace) (*models.Member, bool) {
		var memberFound models.Member
		for _, memberSearch := range namespace.Members {
			if memberSearch.ID == member.ID {
				memberFound = memberSearch
			}
		}
		if memberFound.ID == "" || memberFound.Type == "" {
			return nil, false
		}

		return &memberFound, true
	}

	memberPassive, _, err := s.store.UserGetByID(ctx, memberID, false)
	if err != nil {
		if err == store.ErrNoDocuments {
			return ErrUserNotFound
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

	namespace, err := s.store.NamespaceGet(ctx, tenantID)
	if err != nil {
		if err == store.ErrNoDocuments {
			return ErrNamespaceNotFound
		}

		return err
	}

	passiveMemberFound, ok := findMemberNamespace(memberPassive, namespace)
	if !ok {
		return ErrNamespaceMemberNotFound
	}
	activeMemberFound, ok := findMemberNamespace(memberActive, namespace)
	if !ok {
		return ErrNamespaceMemberNotFound
	}

	if activeMemberFound.Type == passiveMemberFound.Type {
		return ErrForbidden
	}

	if !guard.EvaluateSubject(ctx, s.store, tenantID, userID, memberNewType) {
		return ErrForbidden
	}

	return s.store.NamespaceEditMember(ctx, tenantID, memberPassive.ID, memberNewType)
}

func (s *service) EditSessionRecordStatus(ctx context.Context, sessionRecord bool, tenantID string) error {
	return s.store.NamespaceSetSessionRecord(ctx, sessionRecord, tenantID)
}

func (s *service) GetSessionRecord(ctx context.Context, tenantID string) (bool, error) {
	if _, err := s.store.NamespaceGet(ctx, tenantID); err != nil {
		if err == store.ErrNoDocuments {
			return false, ErrNamespaceNotFound
		}

		return false, err
	}

	return s.store.NamespaceGetSessionRecord(ctx, tenantID)
}
