package services

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/api/store"
	storemock "github.com/shellhub-io/shellhub/api/store/mocks"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	storecache "github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// A connection is personal: GetConnection resolves it scoped to both the
// namespace and the caller (InNamespace + OwnedBy). When the store finds no such
// row (because it belongs to another user, or doesn't exist) the service returns
// NotFound, never leaking another user's connection.
func TestGetConnection(t *testing.T) {
	storeMock := storemock.NewMockStore(t)
	queryOptionsMock := storemock.NewMockQueryOptions(t)
	storeMock.On("Options").Return(queryOptionsMock)

	ctx := context.TODO()

	const (
		tenantID = "00000000-0000-4000-0000-000000000000"
		userID   = "60fb0632538a82e62c2c40a1"
		connID   = "11111111-1111-4111-8111-111111111111"
	)

	owned := &models.Connection{
		ID:       connID,
		TenantID: tenantID,
		OwnerID:  userID,
		Label:    "db-primary",
		Kind:     models.ConnectionKindExternal,
		Host:     "10.0.0.5",
		Port:     22,
	}

	type Expected struct {
		connection *models.Connection
		err        error
	}

	cases := []struct {
		description   string
		req           *requests.ConnectionGet
		requiredMocks func()
		expected      Expected
	}{
		{
			description: "fail with not found when the connection is not the caller's",
			req:         &requests.ConnectionGet{TenantID: tenantID, UserID: userID, ID: connID},
			requiredMocks: func() {
				queryOptionsMock.On("InNamespace", tenantID).Return(nil).Once()
				queryOptionsMock.On("OwnedBy", userID).Return(nil).Once()
				storeMock.
					On("ConnectionResolve", ctx, store.ConnectionIDResolver, connID,
						mock.MatchedBy(func(opts []store.QueryOption) bool { return len(opts) == 2 })).
					Return(nil, store.ErrNoDocuments).
					Once()
			},
			expected: Expected{
				connection: nil,
				err:        NewErrConnectionNotFound(connID, store.ErrNoDocuments),
			},
		},
		{
			description: "success when the connection is owned by the caller",
			req:         &requests.ConnectionGet{TenantID: tenantID, UserID: userID, ID: connID},
			requiredMocks: func() {
				queryOptionsMock.On("InNamespace", tenantID).Return(nil).Once()
				queryOptionsMock.On("OwnedBy", userID).Return(nil).Once()
				storeMock.
					On("ConnectionResolve", ctx, store.ConnectionIDResolver, connID,
						mock.MatchedBy(func(opts []store.QueryOption) bool { return len(opts) == 2 })).
					Return(owned, nil).
					Once()
			},
			expected: Expected{connection: owned, err: nil},
		},
	}

	s := NewService(store.Store(storeMock), privateKey, publicKey, storecache.NewNullCache(), clientMock)

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			connection, err := s.GetConnection(ctx, tc.req)
			assert.Equal(t, tc.expected, Expected{connection, err})
		})
	}

	storeMock.AssertExpectations(t)
}

// The target kind is fixed at creation. An update that flips an external host to a
// device (or vice-versa) is rejected before any write, so a saved target can't
// change shape under a stable id.
func TestUpdateConnectionRejectsKindChange(t *testing.T) {
	storeMock := storemock.NewMockStore(t)
	queryOptionsMock := storemock.NewMockQueryOptions(t)
	storeMock.On("Options").Return(queryOptionsMock)

	ctx := context.TODO()

	const (
		tenantID = "00000000-0000-4000-0000-000000000000"
		userID   = "60fb0632538a82e62c2c40a1"
		connID   = "11111111-1111-4111-8111-111111111111"
	)

	owned := &models.Connection{
		ID:       connID,
		TenantID: tenantID,
		OwnerID:  userID,
		Label:    "db-primary",
		Kind:     models.ConnectionKindExternal,
		Host:     "10.0.0.5",
		Port:     22,
	}

	queryOptionsMock.On("InNamespace", tenantID).Return(nil).Once()
	queryOptionsMock.On("OwnedBy", userID).Return(nil).Once()
	storeMock.
		On("ConnectionResolve", ctx, store.ConnectionIDResolver, connID,
			mock.MatchedBy(func(opts []store.QueryOption) bool { return len(opts) == 2 })).
		Return(owned, nil).
		Once()

	s := NewService(store.Store(storeMock), privateKey, publicKey, storecache.NewNullCache(), clientMock)

	// Same id, but now pointed at a device.
	connection, err := s.UpdateConnection(ctx, &requests.ConnectionUpdate{
		TenantID:  tenantID,
		UserID:    userID,
		ID:        connID,
		Label:     "db-primary",
		Kind:      "device",
		DeviceUID: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
	})

	assert.Nil(t, connection)
	assert.Equal(t, NewErrConnectionKindImmutable(), err)

	// The kind guard runs before any device validation or write.
	storeMock.AssertNotCalled(t, "DeviceResolve", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	storeMock.AssertNotCalled(t, "ConnectionUpdate", mock.Anything, mock.Anything)
	storeMock.AssertExpectations(t)
}
