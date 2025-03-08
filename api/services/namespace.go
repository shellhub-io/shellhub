package services

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
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
	return nil, nil
}

func (s *service) ListNamespaces(ctx context.Context, req *requests.NamespaceList) ([]models.Namespace, int, error) {
	return nil, 0, nil
}

// GetNamespace gets a namespace.
//
// It receives a context, used to "control" the request flow and the tenant ID from models.Namespace.
//
// GetNamespace returns a models.Namespace and an error. When error is not nil, the models.Namespace is nil.
func (s *service) GetNamespace(ctx context.Context, tenantID string) (*models.Namespace, error) {
	return nil, nil
}

// DeleteNamespace deletes a namespace.
//
// It receives a context, used to "control" the request flow and the tenant ID from models.Namespace.
//
// When cloud and billing is enabled, it will try to delete the namespace's billing information from the billing
// service if it exists.
func (s *service) DeleteNamespace(ctx context.Context, tenantID string) error {
	return nil
}

func (s *service) EditNamespace(ctx context.Context, req *requests.NamespaceEdit) (*models.Namespace, error) {
	return nil, nil
}

// EditSessionRecordStatus defines if the sessions will be recorded.
//
// It receives a context, used to "control" the request flow, a boolean to define if the sessions will be recorded and
// the tenant ID from models.Namespace.
func (s *service) EditSessionRecordStatus(ctx context.Context, sessionRecord bool, tenantID string) error {
	return nil
}

// GetSessionRecord gets the session record data.
//
// It receives a context, used to "control" the request flow, the tenant ID from models.Namespace.
//
// GetSessionRecord returns a boolean indicating the session record status and an error. When error is not nil,
// the boolean is false.
func (s *service) GetSessionRecord(ctx context.Context, tenantID string) (bool, error) {
	return false, nil
}
