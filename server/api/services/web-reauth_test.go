package services

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/requests"
	mockcache "github.com/shellhub-io/shellhub/pkg/cache/mocks"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/store"
	storemock "github.com/shellhub-io/shellhub/server/api/store/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestWebReauthVerify(t *testing.T) {
	ctx := context.TODO()

	const (
		tenantID    = "00000000-0000-4000-0000-000000000000"
		userID      = "00000000-0000-0000-0000-00000000000a"
		otherUserID = "00000000-0000-0000-0000-00000000000b"
		fingerprint = "SHA256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	)

	const hash = "$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YWQCIa2UYuFV4OJby7Yi"
	user := &models.User{ID: userID, Password: models.UserPassword{Hash: hash}}

	cases := []struct {
		description  string
		req          *requests.WebReauthVerify
		requireMocks func(storeMock *storemock.MockStore, queryOptionsMock *storemock.MockQueryOptions)
		expectedErr  bool
	}{
		{
			description: "refreshes the identity's re-auth window on a correct password",
			req:         &requests.WebReauthVerify{TenantID: tenantID, UserID: userID, Password: "correct-horse", Fingerprint: fingerprint},
			requireMocks: func(storeMock *storemock.MockStore, queryOptionsMock *storemock.MockQueryOptions) {
				storeMock.On("UserResolve", ctx, store.UserIDResolver, userID).Return(user, nil).Once()
				hashMock.On("CompareWith", "correct-horse", hash).Return(true).Once()
				storeMock.On("SSHIdentityResolve", ctx, mock.Anything, store.SSHIdentityFingerprintResolver, fingerprint).
					Return(&models.SSHIdentity{PrincipalID: userID, Fingerprint: fingerprint}, nil).Once()
				storeMock.On("WithTransaction", mock.Anything, mock.AnythingOfType("store.TransactionCb")).
					Return(func(ctx context.Context, cb store.TransactionCb) error { return cb(ctx) }).Once()
				storeMock.On("SSHIdentityTouchReauth", mock.Anything, tenantID, fingerprint).Return(nil).Once()
			},
			expectedErr: false,
		},
		{
			description: "rejects a wrong password without stamping",
			req:         &requests.WebReauthVerify{TenantID: tenantID, UserID: userID, Password: "wrong", Fingerprint: fingerprint},
			requireMocks: func(storeMock *storemock.MockStore, _ *storemock.MockQueryOptions) {
				storeMock.On("UserResolve", ctx, store.UserIDResolver, userID).Return(user, nil).Once()
				hashMock.On("CompareWith", "wrong", hash).Return(false).Once()
			},
			expectedErr: true,
		},
		{
			description: "fails when the user does not exist",
			req:         &requests.WebReauthVerify{TenantID: tenantID, UserID: userID, Password: "correct-horse", Fingerprint: fingerprint},
			requireMocks: func(storeMock *storemock.MockStore, _ *storemock.MockQueryOptions) {
				storeMock.On("UserResolve", ctx, store.UserIDResolver, userID).Return(nil, store.ErrNoDocuments).Once()
			},
			expectedErr: true,
		},
		{
			description: "rejects a fingerprint owned by another member",
			req:         &requests.WebReauthVerify{TenantID: tenantID, UserID: userID, Password: "correct-horse", Fingerprint: fingerprint},
			requireMocks: func(storeMock *storemock.MockStore, queryOptionsMock *storemock.MockQueryOptions) {
				storeMock.On("UserResolve", ctx, store.UserIDResolver, userID).Return(user, nil).Once()
				hashMock.On("CompareWith", "correct-horse", hash).Return(true).Once()
				storeMock.On("SSHIdentityResolve", ctx, mock.Anything, store.SSHIdentityFingerprintResolver, fingerprint).
					Return(&models.SSHIdentity{PrincipalID: otherUserID, Fingerprint: fingerprint}, nil).Once()
			},
			expectedErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			clockMock.On("Now").Return(now)
			storeMock := new(storemock.MockStore)
			queryOptionsMock := new(storemock.MockQueryOptions)
			storeMock.On("Options").Return(queryOptionsMock).Maybe()
			tc.requireMocks(storeMock, queryOptionsMock)

			service := NewService(store.Store(storeMock), privateKey, publicKey, new(mockcache.MockCache))

			err := service.WebReauthVerify(ctx, tc.req)
			if tc.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			storeMock.AssertExpectations(t)
		})
	}
}

// The step-up is what releases a held login: proving the factor and letting the
// login through happen in one call, so they cannot drift apart.
func TestStampWebReauthReleasesTheHeldLogin(t *testing.T) {
	ctx := context.TODO()

	const (
		tenantID    = "00000000-0000-4000-0000-000000000000"
		userID      = "00000000-0000-0000-0000-00000000000a"
		fingerprint = "SHA256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
		code        = "WXYZ2K7Q"
	)

	reauthApproval := &models.SSHApproval{
		Code:        code,
		TenantID:    tenantID,
		Kind:        models.SSHApprovalReauth,
		Fingerprint: fingerprint,
		State:       models.SSHApprovalPending,
	}

	cases := []struct {
		description  string
		approval     *models.SSHApproval
		approvalErr  error
		claimed      bool
		expectDecide bool
		expectedErr  bool
	}{
		{
			description:  "releases the login the step-up was for",
			approval:     reauthApproval,
			claimed:      true,
			expectDecide: true,
			expectedErr:  false,
		},
		{
			description: "refuses an approval for a different key",
			approval: &models.SSHApproval{
				Code: code, TenantID: tenantID, Kind: models.SSHApprovalReauth,
				Fingerprint: "SHA256:someone-else", State: models.SSHApprovalPending,
			},
			expectedErr: true,
		},
		{
			description: "refuses an identity approval",
			approval: &models.SSHApproval{
				Code: code, TenantID: tenantID, Kind: models.SSHApprovalIdentity,
				Fingerprint: fingerprint, State: models.SSHApprovalPending,
			},
			expectedErr: true,
		},
		{
			description:  "fails when the login was already decided",
			approval:     reauthApproval,
			claimed:      false,
			expectDecide: true,
			expectedErr:  true,
		},
		{
			description: "fails when the code is unknown or expired",
			approvalErr: store.ErrNoDocuments,
			expectedErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			clockMock.On("Now").Return(now)
			storeMock := new(storemock.MockStore)
			queryOptionsMock := new(storemock.MockQueryOptions)
			storeMock.On("Options").Return(queryOptionsMock).Maybe()

			storeMock.On("SSHIdentityResolve", ctx, mock.Anything, store.SSHIdentityFingerprintResolver, fingerprint).
				Return(&models.SSHIdentity{PrincipalID: userID, Fingerprint: fingerprint}, nil).Once()
			storeMock.On("WithTransaction", mock.Anything, mock.AnythingOfType("store.TransactionCb")).
				Return(func(ctx context.Context, cb store.TransactionCb) error { return cb(ctx) }).Once()
			storeMock.On("SSHIdentityTouchReauth", mock.Anything, tenantID, fingerprint).Return(nil).Once()
			storeMock.On("SSHApprovalGet", mock.Anything, code, now).Return(tc.approval, tc.approvalErr).Once()

			if tc.expectDecide {
				storeMock.On("SSHApprovalDecide", mock.Anything, code, models.SSHApprovalConfirmed, userID, now).
					Return(tc.claimed, nil).Once()
			}

			err := StampWebReauth(ctx, store.Store(storeMock), &requests.WebReauthVerify{
				TenantID:     tenantID,
				UserID:       userID,
				Fingerprint:  fingerprint,
				ApprovalCode: code,
			})

			if tc.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			storeMock.AssertExpectations(t)
		})
	}
}
