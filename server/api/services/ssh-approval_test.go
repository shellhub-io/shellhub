package services

import (
	"context"
	"regexp"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	cachemock "github.com/shellhub-io/shellhub/pkg/cache/mocks"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/store"
	storemock "github.com/shellhub-io/shellhub/server/api/store/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreateSSHApproval(t *testing.T) {
	req := &requests.SSHApprovalCreate{
		SessionUID:  "session1",
		SSHID:       "root@namespace.device",
		TenantID:    "tenant1",
		DeviceUID:   "device1",
		DeviceName:  "device",
		Username:    "root",
		IPAddress:   "10.0.0.1",
		Kind:        models.SSHApprovalIdentity,
		Fingerprint: "SHA256:abc",
		Data:        []byte("ssh-ed25519 AAAA"),
	}

	cacheMock := new(cachemock.MockCache)
	storeMock := new(storemock.MockStore)

	clockMock.On("Now").Return(now)

	storeMock.
		On("SSHApprovalCreate", mock.Anything, mock.MatchedBy(func(approval *models.SSHApproval) bool {
			return regexp.MustCompile(`^[2-9A-HJKMNP-TV-Z]{8}$`).MatchString(approval.Code) &&
				approval.SessionUID == "session1" &&
				approval.TenantID == "tenant1" &&
				approval.Kind == models.SSHApprovalIdentity &&
				approval.Fingerprint == "SHA256:abc" &&
				approval.State == models.SSHApprovalPending &&
				approval.RequestedAt.Equal(now) &&
				approval.ExpiresAt.Equal(now.Add(sshApprovalTTL))
		})).
		Return(nil).
		Once()

	service := NewService(storeMock, privateKey, publicKey, cacheMock)

	approval, err := service.CreateSSHApproval(context.TODO(), req)
	require.NoError(t, err)
	require.Regexp(t, `^[2-9A-HJKMNP-TV-Z]{8}$`, approval.Code)
	require.Equal(t, int(sshApprovalTTL.Seconds()), approval.ExpiresIn)

	storeMock.AssertExpectations(t)
}

func TestGetSSHApprovalStatus(t *testing.T) {
	type Expected struct {
		status *models.SSHApprovalStatus
		err    error
	}

	cases := []struct {
		description   string
		req           *requests.SSHApprovalStatus
		requiredMocks func(storeMock *storemock.MockStore)
		expected      Expected
	}{
		{
			description: "fails when the code is unknown or expired",
			req:         &requests.SSHApprovalStatus{Code: "WXYZ2K7Q"},
			requiredMocks: func(storeMock *storemock.MockStore) {
				storeMock.
					On("SSHApprovalGet", mock.Anything, "WXYZ2K7Q", now).
					Return(nil, store.ErrNoDocuments).
					Once()
			},
			expected: Expected{status: nil, err: NewErrSSHApprovalCodeNotFound("WXYZ2K7Q", store.ErrNoDocuments)},
		},
		{
			description: "reports pending without waiting when not asked to wait",
			req:         &requests.SSHApprovalStatus{Code: "WXYZ2K7Q"},
			requiredMocks: func(storeMock *storemock.MockStore) {
				storeMock.
					On("SSHApprovalGet", mock.Anything, "WXYZ2K7Q", now).
					Return(&models.SSHApproval{State: models.SSHApprovalPending}, nil).
					Once()
			},
			expected: Expected{status: &models.SSHApprovalStatus{State: models.SSHApprovalPending}, err: nil},
		},
		{
			description: "waits through a pending read and answers on the decision",
			req:         &requests.SSHApprovalStatus{Code: "WXYZ2K7Q", Wait: true},
			requiredMocks: func(storeMock *storemock.MockStore) {
				storeMock.
					On("SSHApprovalGet", mock.Anything, "WXYZ2K7Q", now).
					Return(&models.SSHApproval{State: models.SSHApprovalPending}, nil).
					Once()
				storeMock.
					On("SSHApprovalGet", mock.Anything, "WXYZ2K7Q", now).
					Return(&models.SSHApproval{State: models.SSHApprovalConfirmed, DecidedBy: "owner1"}, nil).
					Once()
			},
			expected: Expected{
				status: &models.SSHApprovalStatus{State: models.SSHApprovalConfirmed, UserID: "owner1"},
				err:    nil,
			},
		},
		{
			description: "returns the approver once decided",
			req:         &requests.SSHApprovalStatus{Code: "WXYZ2K7Q"},
			requiredMocks: func(storeMock *storemock.MockStore) {
				storeMock.
					On("SSHApprovalGet", mock.Anything, "WXYZ2K7Q", now).
					Return(&models.SSHApproval{State: models.SSHApprovalConfirmed, DecidedBy: "owner1"}, nil).
					Once()
			},
			expected: Expected{
				status: &models.SSHApprovalStatus{State: models.SSHApprovalConfirmed, UserID: "owner1"},
				err:    nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			storeMock := new(storemock.MockStore)
			clockMock.On("Now").Return(now)
			tc.requiredMocks(storeMock)

			service := NewService(storeMock, privateKey, publicKey, new(cachemock.MockCache))

			status, err := service.GetSSHApprovalStatus(context.TODO(), tc.req)
			require.Equal(tt, tc.expected.err, err)
			require.Equal(tt, tc.expected.status, status)

			storeMock.AssertExpectations(tt)
		})
	}
}

func TestGetSSHApproval(t *testing.T) {
	namespace := &models.Namespace{
		Name:     "namespace1",
		TenantID: "tenant1",
		Members:  []models.Member{{ID: "owner1", Role: authorizer.RoleOwner}},
	}

	pending := &models.SSHApproval{
		SSHID:       "root@namespace1.device",
		TenantID:    "tenant1",
		DeviceName:  "device",
		Username:    "root",
		IPAddress:   "10.0.0.1",
		Fingerprint: "SHA256:fingerprint",
		Kind:        models.SSHApprovalIdentity,
		State:       models.SSHApprovalPending,
		ExpiresAt:   now.Add(sshApprovalTTL),
	}

	type Expected struct {
		request *models.SSHApprovalRequest
		err     error
	}

	cases := []struct {
		description   string
		userID        string
		code          string
		requiredMocks func(storeMock *storemock.MockStore)
		expected      Expected
	}{
		{
			description: "fails when the code is unknown or expired",
			userID:      "owner1",
			code:        "WXYZ2K7Q",
			requiredMocks: func(storeMock *storemock.MockStore) {
				storeMock.
					On("SSHApprovalGet", mock.Anything, "WXYZ2K7Q", now).
					Return(nil, store.ErrNoDocuments).
					Once()
			},
			expected: Expected{request: nil, err: NewErrSSHApprovalCodeNotFound("WXYZ2K7Q", store.ErrNoDocuments)},
		},
		{
			description: "fails when the target namespace does not exist",
			userID:      "owner1",
			code:        "WXYZ2K7Q",
			requiredMocks: func(storeMock *storemock.MockStore) {
				storeMock.
					On("SSHApprovalGet", mock.Anything, "WXYZ2K7Q", now).
					Return(pending, nil).
					Once()
				storeMock.
					On("NamespaceResolve", mock.Anything, store.NamespaceTenantIDResolver, "tenant1").
					Return(nil, store.ErrNoDocuments).
					Once()
			},
			expected: Expected{request: nil, err: NewErrSSHApprovalCodeNotFound("WXYZ2K7Q", store.ErrNoDocuments)},
		},
		{
			description: "hides the request from a non-member, without confirming the code exists",
			userID:      "intruder",
			code:        "WXYZ2K7Q",
			requiredMocks: func(storeMock *storemock.MockStore) {
				storeMock.
					On("SSHApprovalGet", mock.Anything, "WXYZ2K7Q", now).
					Return(pending, nil).
					Once()
				storeMock.
					On("NamespaceResolve", mock.Anything, store.NamespaceTenantIDResolver, "tenant1").
					Return(namespace, nil).
					Once()
			},
			expected: Expected{request: nil, err: NewErrSSHApprovalCodeNotFound("WXYZ2K7Q", nil)},
		},
		{
			description: "returns the details to a member, naming the target namespace",
			userID:      "owner1",
			code:        "WXYZ2K7Q",
			requiredMocks: func(storeMock *storemock.MockStore) {
				storeMock.
					On("SSHApprovalGet", mock.Anything, "WXYZ2K7Q", now).
					Return(pending, nil).
					Once()
				storeMock.
					On("NamespaceResolve", mock.Anything, store.NamespaceTenantIDResolver, "tenant1").
					Return(namespace, nil).
					Once()
			},
			expected: Expected{
				request: &models.SSHApprovalRequest{
					SSHID:       "root@namespace1.device",
					DeviceName:  "device",
					Username:    "root",
					IPAddress:   "10.0.0.1",
					State:       models.SSHApprovalPending,
					Code:        "WXYZ2K7Q",
					Fingerprint: "SHA256:fingerprint",
					Kind:        models.SSHApprovalIdentity,
					ExpiresIn:   int(sshApprovalTTL.Seconds()),
					Namespace:   "namespace1",
				},
				err: nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			storeMock := new(storemock.MockStore)
			clockMock.On("Now").Return(now)
			tc.requiredMocks(storeMock)

			service := NewService(storeMock, privateKey, publicKey, new(cachemock.MockCache))

			request, err := service.GetSSHApproval(context.TODO(), tc.userID, tc.code)
			require.Equal(tt, tc.expected.err, err)
			require.Equal(tt, tc.expected.request, request)

			storeMock.AssertExpectations(tt)
		})
	}
}

// TestSSHApprovalDecideAuthorization covers who may decide, independent of what
// the decision then does. It runs through reject, which carries no kind-specific
// rule; the effect of each kind is covered separately below.
func TestSSHApprovalDecideAuthorization(t *testing.T) {
	namespace := &models.Namespace{
		Name:     "namespace1",
		TenantID: "tenant1",
		Members: []models.Member{
			{ID: "owner1", Role: authorizer.RoleOwner},
			{ID: "observer1", Role: authorizer.RoleObserver},
		},
	}

	pending := &models.SSHApproval{
		SessionUID:  "session1",
		TenantID:    "tenant1",
		Kind:        models.SSHApprovalReauth,
		Fingerprint: "SHA256:abc",
		State:       models.SSHApprovalPending,
	}

	cases := []struct {
		description   string
		userID        string
		code          string
		requiredMocks func(storeMock *storemock.MockStore)
		expectedErr   error
	}{
		{
			description:   "fails when the code is not well-formed",
			userID:        "owner1",
			code:          "00000000",
			requiredMocks: func(_ *storemock.MockStore) {},
			expectedErr:   NewErrSSHApprovalCodeNotFound("00000000", nil),
		},
		{
			description: "fails when the code is unknown or expired",
			userID:      "owner1",
			code:        "WXYZ2K7Q",
			requiredMocks: func(storeMock *storemock.MockStore) {
				storeMock.
					On("SSHApprovalGet", mock.Anything, "WXYZ2K7Q", now).
					Return(nil, store.ErrNoDocuments).
					Once()
			},
			expectedErr: NewErrSSHApprovalCodeNotFound("WXYZ2K7Q", store.ErrNoDocuments),
		},
		{
			description: "fails when the target namespace does not exist",
			userID:      "owner1",
			code:        "WXYZ2K7Q",
			requiredMocks: func(storeMock *storemock.MockStore) {
				storeMock.On("SSHApprovalGet", mock.Anything, "WXYZ2K7Q", now).Return(pending, nil).Once()
				storeMock.
					On("NamespaceResolve", mock.Anything, store.NamespaceTenantIDResolver, "tenant1").
					Return(nil, store.ErrNoDocuments).
					Once()
			},
			expectedErr: NewErrNamespaceNotFound("tenant1", store.ErrNoDocuments),
		},
		{
			description: "fails when the user is not a member of the namespace",
			userID:      "intruder",
			code:        "WXYZ2K7Q",
			requiredMocks: func(storeMock *storemock.MockStore) {
				storeMock.On("SSHApprovalGet", mock.Anything, "WXYZ2K7Q", now).Return(pending, nil).Once()
				storeMock.
					On("NamespaceResolve", mock.Anything, store.NamespaceTenantIDResolver, "tenant1").
					Return(namespace, nil).
					Once()
			},
			expectedErr: NewErrNamespaceMemberNotFound("intruder", nil),
		},
		{
			description: "fails when the member cannot approve sessions",
			userID:      "observer1",
			code:        "WXYZ2K7Q",
			requiredMocks: func(storeMock *storemock.MockStore) {
				storeMock.On("SSHApprovalGet", mock.Anything, "WXYZ2K7Q", now).Return(pending, nil).Once()
				storeMock.
					On("NamespaceResolve", mock.Anything, store.NamespaceTenantIDResolver, "tenant1").
					Return(namespace, nil).
					Once()
			},
			expectedErr: NewErrRoleForbidden(),
		},
		{
			description: "fails when the code was already decided",
			userID:      "owner1",
			code:        "WXYZ2K7Q",
			requiredMocks: func(storeMock *storemock.MockStore) {
				storeMock.On("SSHApprovalGet", mock.Anything, "WXYZ2K7Q", now).Return(pending, nil).Once()
				storeMock.
					On("NamespaceResolve", mock.Anything, store.NamespaceTenantIDResolver, "tenant1").
					Return(namespace, nil).
					Once()
				storeMock.
					On("WithTransaction", mock.Anything, mock.AnythingOfType("store.TransactionCb")).
					Return(func(ctx context.Context, cb store.TransactionCb) error { return cb(ctx) }).
					Once()
				storeMock.
					On("SSHApprovalDecide", mock.Anything, "WXYZ2K7Q", models.SSHApprovalRejected, "owner1", now).
					Return(false, nil).
					Once()
			},
			expectedErr: NewErrSSHApprovalCodeNotFound("WXYZ2K7Q", nil),
		},
		{
			description: "decides and binds the decider",
			userID:      "owner1",
			code:        "WXYZ2K7Q",
			requiredMocks: func(storeMock *storemock.MockStore) {
				storeMock.On("SSHApprovalGet", mock.Anything, "WXYZ2K7Q", now).Return(pending, nil).Once()
				storeMock.
					On("NamespaceResolve", mock.Anything, store.NamespaceTenantIDResolver, "tenant1").
					Return(namespace, nil).
					Once()
				storeMock.
					On("WithTransaction", mock.Anything, mock.AnythingOfType("store.TransactionCb")).
					Return(func(ctx context.Context, cb store.TransactionCb) error { return cb(ctx) }).
					Once()
				storeMock.
					On("SSHApprovalDecide", mock.Anything, "WXYZ2K7Q", models.SSHApprovalRejected, "owner1", now).
					Return(true, nil).
					Once()
			},
			expectedErr: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			storeMock := new(storemock.MockStore)
			clockMock.On("Now").Return(now)
			tc.requiredMocks(storeMock)

			service := NewService(storeMock, privateKey, publicKey, new(cachemock.MockCache))

			err := service.RejectSSHApproval(context.TODO(), tc.userID, &requests.SSHApprovalReject{Code: tc.code})
			require.Equal(tt, tc.expectedErr, err)

			storeMock.AssertExpectations(tt)
		})
	}
}

// A re-auth is satisfied by proving a factor on the step-up route, which
// releases the login there. A bare confirm must not be a way around it.
func TestConfirmSSHApprovalRefusesReauthWithoutAFactor(t *testing.T) {
	namespace := &models.Namespace{
		Name:     "namespace1",
		TenantID: "tenant1",
		Members:  []models.Member{{ID: "owner1", Role: authorizer.RoleOwner}},
	}

	storeMock := new(storemock.MockStore)
	clockMock.On("Now").Return(now)

	storeMock.
		On("SSHApprovalGet", mock.Anything, "WXYZ2K7Q", now).
		Return(&models.SSHApproval{
			TenantID:    "tenant1",
			Kind:        models.SSHApprovalReauth,
			Fingerprint: "SHA256:abc",
			State:       models.SSHApprovalPending,
		}, nil).
		Once()
	storeMock.
		On("NamespaceResolve", mock.Anything, store.NamespaceTenantIDResolver, "tenant1").
		Return(namespace, nil).
		Once()

	service := NewService(storeMock, privateKey, publicKey, new(cachemock.MockCache))

	err := service.ConfirmSSHApproval(context.TODO(), "owner1", &requests.SSHApprovalConfirm{Code: "WXYZ2K7Q"})
	require.Error(t, err)

	storeMock.AssertExpectations(t)
	storeMock.AssertNotCalled(t, "SSHApprovalDecide", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	storeMock.AssertNotCalled(t, "SSHIdentityTouchReauth", mock.Anything, mock.Anything, mock.Anything)
}

// Confirming an identity approval binds the presented key to the approver.
func TestConfirmSSHApprovalIdentity(t *testing.T) {
	namespace := &models.Namespace{
		Name:     "namespace1",
		TenantID: "00000000-0000-4000-0000-000000000000",
		Members:  []models.Member{{ID: "owner1", Role: authorizer.RoleOwner}},
	}

	storeMock := new(storemock.MockStore)
	queryOptionsMock := new(storemock.MockQueryOptions)
	storeMock.On("Options").Return(queryOptionsMock).Maybe()

	clockMock.On("Now").Return(now)

	storeMock.
		On("SSHApprovalGet", mock.Anything, "WXYZ2K7Q", now).
		Return(&models.SSHApproval{
			TenantID:    namespace.TenantID,
			Kind:        models.SSHApprovalIdentity,
			Fingerprint: "SHA256:abc",
			Data:        []byte("ssh-ed25519 AAAA"),
			State:       models.SSHApprovalPending,
		}, nil).
		Once()
	storeMock.
		On("NamespaceResolve", mock.Anything, store.NamespaceTenantIDResolver, namespace.TenantID).
		Return(namespace, nil).
		Once()
	storeMock.
		On("WithTransaction", mock.Anything, mock.AnythingOfType("store.TransactionCb")).
		Return(func(ctx context.Context, cb store.TransactionCb) error { return cb(ctx) }).
		Once()
	storeMock.
		On("SSHApprovalDecide", mock.Anything, "WXYZ2K7Q", models.SSHApprovalConfirmed, "owner1", now).
		Return(true, nil).
		Once()

	storeMock.
		On("SSHIdentityResolve", mock.Anything, mock.Anything, store.SSHIdentityFingerprintResolver, "SHA256:abc").
		Return(nil, store.ErrNoDocuments).
		Once()
	storeMock.
		On("SSHIdentityCreate", mock.Anything, mock.MatchedBy(func(identity *models.SSHIdentity) bool {
			return identity.PrincipalID == "owner1" && identity.Fingerprint == "SHA256:abc"
		})).
		Return("id1", nil).
		Once()
	storeMock.
		On("SSHIdentityTouchLastUsed", mock.Anything, namespace.TenantID, "SHA256:abc").
		Return(nil).
		Once()

	service := NewService(storeMock, privateKey, publicKey, new(cachemock.MockCache))

	err := service.ConfirmSSHApproval(context.TODO(), "owner1", &requests.SSHApprovalConfirm{Code: "WXYZ2K7Q"})
	require.NoError(t, err)

	storeMock.AssertExpectations(t)
}

// Rejecting decides the approval without touching any identity.
func TestRejectSSHApproval(t *testing.T) {
	namespace := &models.Namespace{
		Name:     "namespace1",
		TenantID: "tenant1",
		Members:  []models.Member{{ID: "owner1", Role: authorizer.RoleOwner}},
	}

	storeMock := new(storemock.MockStore)

	clockMock.On("Now").Return(now)

	storeMock.
		On("SSHApprovalGet", mock.Anything, "WXYZ2K7Q", now).
		Return(&models.SSHApproval{
			TenantID:    "tenant1",
			Kind:        models.SSHApprovalIdentity,
			Fingerprint: "SHA256:abc",
			State:       models.SSHApprovalPending,
		}, nil).
		Once()
	storeMock.
		On("NamespaceResolve", mock.Anything, store.NamespaceTenantIDResolver, "tenant1").
		Return(namespace, nil).
		Once()
	storeMock.
		On("WithTransaction", mock.Anything, mock.AnythingOfType("store.TransactionCb")).
		Return(func(ctx context.Context, cb store.TransactionCb) error { return cb(ctx) }).
		Once()
	storeMock.
		On("SSHApprovalDecide", mock.Anything, "WXYZ2K7Q", models.SSHApprovalRejected, "owner1", now).
		Return(true, nil).
		Once()

	service := NewService(storeMock, privateKey, publicKey, new(cachemock.MockCache))

	err := service.RejectSSHApproval(context.TODO(), "owner1", &requests.SSHApprovalReject{Code: "WXYZ2K7Q"})
	require.NoError(t, err)

	storeMock.AssertExpectations(t)
}
