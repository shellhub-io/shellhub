package services

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"

	storecache "github.com/shellhub-io/shellhub/api/cache"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/mocks"
	"github.com/shellhub-io/shellhub/pkg/authorizer"
	"github.com/stretchr/testify/assert"
)

func TestCheckPermission(t *testing.T) {
	mock := &mocks.Store{}

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)

	s := NewService(store.Store(mock), privateKey, &privateKey.PublicKey, storecache.NewNullCache(), clientMock, nil)

	tests := []struct {
		description   string
		userType      string
		action        int
		callback      error
		requiredMocks func()
		expected      error
	}{
		{
			description: "Fail when the user does not have the permission to execute an action",
			userType:    authorizer.MemberTypeObserver,
			action:      authorizer.Actions.Namespace.Rename,
			callback:    nil,
			requiredMocks: func() {
			},
			expected: ErrForbidden,
		},
		{
			description: "Success when the user has permission to execute an action",
			userType:    authorizer.MemberTypeOwner,
			action:      authorizer.Actions.Namespace.Rename,
			callback:    nil,
			requiredMocks: func() {
			},
			expected: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()
			err := s.CheckPermission(tc.userType, tc.action, func() error {
				return tc.callback
			})

			assert.Equal(t, tc.expected, err)
		})
	}

	mock.AssertExpectations(t)
}
