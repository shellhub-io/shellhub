package nsadm

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"strings"

	uuid "github.com/satori/go.uuid"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"gopkg.in/go-playground/validator.v9"
)

var ErrUnauthorized = errors.New("unauthorized")
var ErrUserNotFound = errors.New("user not found")
var ErrNamespaceNotFound = errors.New("namespace not found")
var ErrDuplicateID = errors.New("user already member of this namespace")
var ErrUserOwner = errors.New("cannot remove this user")

type Service interface {
	ListNamespaces(ctx context.Context, pagination paginator.Query, filterB64 string, export bool) ([]models.Namespace, int, error)
	CreateNamespace(ctx context.Context, namespace *models.Namespace, ownerUsername string) (*models.Namespace, error)
	GetNamespace(ctx context.Context, namespace string) (*models.Namespace, error)
	DeleteNamespace(ctx context.Context, namespace, ownerUsername string) error
	EditNamespace(ctx context.Context, namespace, name, ownerUsername string) (*models.Namespace, error)
	AddNamespaceUser(ctx context.Context, namespace, username, ownerUsername string) (*models.Namespace, error)
	RemoveNamespaceUser(ctx context.Context, namespace, username, ownerUsername string) (*models.Namespace, error)
	ListMembers(ctx context.Context, namespace string) ([]models.Member, error)
	UpdateDataUserSecurity(ctx context.Context, status bool, tenant string) error
	GetDataUserSecurity(ctx context.Context, tenant string) (bool, error)
}

type service struct {
	store store.Store
}

func NewService(store store.Store) Service {
	return &service{store}
}

func (s *service) ListNamespaces(ctx context.Context, pagination paginator.Query, filterB64 string, export bool) ([]models.Namespace, int, error) {
	raw, err := base64.StdEncoding.DecodeString(filterB64)
	if err != nil {
		return nil, 0, err
	}
	var filter []models.Filter

	if err := json.Unmarshal([]byte(raw), &filter); len(raw) > 0 && err != nil {
		return nil, 0, err
	}

	return s.store.ListNamespaces(ctx, pagination, filter, export)
}

func (s *service) CreateNamespace(ctx context.Context, namespace *models.Namespace, ownerID string) (*models.Namespace, error) {
	user, _ := s.store.GetUserByID(ctx, ownerID)
	if user == nil {
		return nil, ErrUnauthorized
	}
	namespace.Name = strings.ToLower(namespace.Name)
	namespace.Owner = user.ID
	members := []string{user.ID}
	namespace.Members = &members
	settings := &models.NamespaceSettings{SessionRecord: true}
	namespace.Settings = settings
	if namespace.TenantID == "" {
		namespace.TenantID = uuid.Must(uuid.NewV4(), nil).String()
	}
	if os.Getenv("SHELLHUB_ENTERPRISE") == "true" {
		namespace.MaxDevices = 3
	} else {
		namespace.MaxDevices = -1
	}

	return s.store.CreateNamespace(ctx, namespace)
}

func (s *service) GetNamespace(ctx context.Context, namespace string) (*models.Namespace, error) {
	return s.store.GetNamespace(ctx, namespace)
}

func (s *service) DeleteNamespace(ctx context.Context, namespace, ownerId string) error {
	ns, _ := s.store.GetNamespace(ctx, namespace)
	if ns != nil {
		user, _ := s.store.GetUserByID(ctx, ownerId)
		if user != nil {
			if ns.Owner == user.ID {
				return s.store.DeleteNamespace(ctx, namespace)
			}
		}
		return ErrUnauthorized
	}
	return ErrNamespaceNotFound
}

func (s *service) ListMembers(ctx context.Context, namespace string) ([]models.Member, error) {
	ns, _ := s.store.GetNamespace(ctx, namespace)
	if ns != nil {
		members := []models.Member{}
		for _, memberID := range ns.Members.(primitive.A) {
			if user, err := s.store.GetUserByID(ctx, memberID.(string)); err == nil {
				member := models.Member{ID: memberID.(string), Name: user.Username}
				members = append(members, member)
			}
		}
		return members, nil
	}
	return []models.Member{}, ErrNamespaceNotFound
}

func (s *service) EditNamespace(ctx context.Context, namespace, name, owner string) (*models.Namespace, error) {
	ns, _ := s.store.GetNamespace(ctx, namespace)
	if ns != nil {
		user, _ := s.store.GetUserByID(ctx, owner)
		if user != nil {
			validate := validator.New()
			name = strings.ToLower(name)
			if ns.Name != name && ns.Owner == user.ID {
				ns.Name = name
				if err := validate.Struct(ns); err == nil {
					return s.store.EditNamespace(ctx, namespace, name)
				}
			}
		}
		return nil, ErrUnauthorized
	}
	return nil, ErrNamespaceNotFound
}

func (s *service) AddNamespaceUser(ctx context.Context, namespace, username, ownerID string) (*models.Namespace, error) {
	ns, _ := s.store.GetNamespace(ctx, namespace)
	if ns != nil {
		if OwnerUser, _ := s.store.GetUserByID(ctx, ownerID); OwnerUser != nil {
			if ns.Owner == OwnerUser.ID {
				if user, _ := s.store.GetUserByUsername(ctx, username); user != nil {
					return s.store.AddNamespaceUser(ctx, namespace, user.ID)
				}
				return nil, ErrUserNotFound
			}
		}
		return nil, ErrUnauthorized
	}
	return nil, ErrNamespaceNotFound
}
func (s *service) RemoveNamespaceUser(ctx context.Context, namespace, username, ownerID string) (*models.Namespace, error) {
	ns, _ := s.store.GetNamespace(ctx, namespace)
	if ns != nil {
		if OwnerUser, _ := s.store.GetUserByID(ctx, ownerID); OwnerUser != nil && OwnerUser.Username != username {
			if ns.Owner == OwnerUser.ID {
				if user, _ := s.store.GetUserByUsername(ctx, username); user != nil {
					if ns, err := s.store.RemoveNamespaceUser(ctx, namespace, user.ID); err == nil {
						return ns, err
					}
				}
				return nil, ErrUserNotFound
			}
		}
		return nil, ErrUnauthorized
	}
	return nil, ErrNamespaceNotFound
}

func (s *service) UpdateDataUserSecurity(ctx context.Context, sessionRecord bool, tenant string) error {
	ns, _ := s.GetNamespace(ctx, tenant)
	if ns != nil {
		return s.store.UpdateDataUserSecurity(ctx, sessionRecord, tenant)
	}
	return ErrUnauthorized
}

func (s *service) GetDataUserSecurity(ctx context.Context, tenant string) (bool, error) {
	ns, _ := s.GetNamespace(ctx, tenant)
	if ns != nil {
		return s.store.GetDataUserSecurity(ctx, tenant)
	}
	return false, ErrUnauthorized
}
