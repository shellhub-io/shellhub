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
	RemoveNamespaceUser(ctx context.Context, tenantID, memberUsername, userID string) (*models.Namespace, error)
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
	if _, err := validator.ValidateStruct(models.Member{Username: memberUsername, Type: memberType}); err != nil {
		return nil, ErrInvalidFormat
	}

	if !guard.EvaluateSubjectType(ctx, s.store, tenantID, userID, memberType) {
		return nil, ErrForbidden
	}

	member, err := s.store.UserGetByUsername(ctx, memberUsername)
	if err != nil {
		if err == store.ErrNoDocuments {
			return nil, ErrUserNotFound
		}

		return nil, err
	}

	return s.store.NamespaceAddMember(ctx, tenantID, member.ID, memberType)
}

func (s *service) RemoveNamespaceUser(ctx context.Context, tenantID, memberUsername, userID string) (*models.Namespace, error) {
	if _, err := validator.ValidateVar(memberUsername, "min=3,max=30,alphanum,ascii"); err != nil { // TODO Remove this static tag string.
		return nil, ErrInvalidFormat
	}

	if !guard.EvaluateSubjectWithUsername(ctx, s.store, tenantID, userID, memberUsername) {
		return nil, ErrForbidden
	}

	user, err := s.store.UserGetByUsername(ctx, memberUsername)
	if err != nil {
		if err == store.ErrNoDocuments {
			return nil, ErrUserNotFound
		}

		return nil, err
	}

	namespace, err := s.GetNamespace(ctx, tenantID)
	if err != nil {
		return nil, ErrNamespaceNotFound
	}

	// You cannot remove the owner from the namespace.
	for _, member := range namespace.Members {
		if member.ID == user.ID && member.Type == authorizer.MemberTypeOwner {
			return nil, ErrForbidden
		}
	}

	return s.store.NamespaceRemoveMember(ctx, tenantID, user.ID)
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
