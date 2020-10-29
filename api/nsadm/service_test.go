package nsadm

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/mocks"
	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestListNamespaces(t *testing.T) {
	mock := &mocks.Store{}
	s := NewService(store.Store(mock))

	ctx := context.TODO()

	namespaces := []models.Namespace{
		{Name: "group1", Owner: "user1", TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713"},
		{Name: "group2", Owner: "user2", TenantID: "a736a52b-5777-4f92-b0b8-e359bf48471i4"},
	}
	query := paginator.Query{Page: 1, PerPage: 10}
	mock.On("ListNamespaces", ctx, query).Return(namespaces, len(namespaces), nil).Once()
	returnedNamespaces, count, err := s.ListNamespaces(ctx, query)
	assert.NoError(t, err)
	assert.Equal(t, namespaces, returnedNamespaces)
	assert.Equal(t, count, len(namespaces))
	mock.AssertExpectations(t)
}

func TestGetNamespace(t *testing.T) {
	mock := &mocks.Store{}
	s := NewService(store.Store(mock))

	ctx := context.TODO()

	namespace := &models.Namespace{Name: "group1", Owner: "user1", TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713"}

	mock.On("GetNamespace", ctx, namespace.TenantID).Return(namespace, nil).Once()

	returnNamespace, err := s.GetNamespace(ctx, namespace.TenantID)
	assert.NoError(t, err)
	assert.Equal(t, namespace, returnNamespace)

	mock.AssertExpectations(t)
}

func TestCreateNamespace(t *testing.T) {
	mock := &mocks.Store{}
	s := NewService(store.Store(mock))

	ctx := context.TODO()

	namespace := &models.Namespace{Name: "group1", Owner: "user1", TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713"}

	mock.On("CreateNamespace", ctx, namespace).Return(namespace, nil).Once()

	returnedNamespace, err := s.CreateNamespace(ctx, namespace)
	assert.NoError(t, err)
	assert.Equal(t, namespace, returnedNamespace)
	mock.AssertExpectations(t)
}

func TestEditNamespace(t *testing.T) {
	mock := &mocks.Store{}
	s := NewService(store.Store(mock))

	ctx := context.TODO()

	namespace := &models.Namespace{Name: "group1", Owner: "user1", TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713"}
	renamedNamespace := &models.Namespace{Name: "group2", Owner: "user1", TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713"}

	mock.On("EditNamespace", ctx, namespace.TenantID, renamedNamespace.Name).Return(nil).Once()
	mock.On("GetNamespace", ctx, namespace.TenantID).Return(namespace, nil).Twice()
	err := s.EditNamespace(ctx, namespace.TenantID, renamedNamespace.Name)
	assert.NoError(t, err)
	returnedNamespace, err := s.GetNamespace(ctx, namespace.TenantID)
	assert.NoError(t, err)
	assert.Equal(t, renamedNamespace, returnedNamespace)
	mock.AssertExpectations(t)
}

func TestDeleteNamespace(t *testing.T) {
	mock := &mocks.Store{}
	s := NewService(store.Store(mock))

	ctx := context.TODO()

	namespace := &models.Namespace{Name: "group1", Owner: "user1", TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713"}

	mock.On("DeleteNamespace", ctx, namespace.TenantID).Return(nil).Once()
	mock.On("GetNamespace", ctx, namespace.TenantID).Return(namespace, nil).Once()

	err := s.DeleteNamespace(ctx, namespace.TenantID)
	assert.NoError(t, err)

	mock.AssertExpectations(t)
}
