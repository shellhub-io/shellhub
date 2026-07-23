package services

import (
	"context"
	"regexp"
	"testing"

	"github.com/shellhub-io/shellhub/api/store"
	storemock "github.com/shellhub-io/shellhub/api/store/mocks"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	cachemock "github.com/shellhub-io/shellhub/pkg/cache/mocks"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreateSSHEnrollment(t *testing.T) {
	req := &requests.SSHEnrollmentCreate{
		SessionUID: "session1",
		SSHID:      "root@namespace.device",
		TenantID:   "tenant1",
		DeviceUID:  "device1",
		DeviceName: "device",
		Username:   "root",
		IPAddress:  "10.0.0.1",
	}

	cacheMock := new(cachemock.MockCache)
	storeMock := new(storemock.MockStore)

	clockMock.On("Now").Return(now)

	cacheMock.
		On("Set", mock.Anything, mock.MatchedBy(func(key string) bool {
			return regexp.MustCompile(`^ssh_enrollment/[2-9A-HJKMNP-TV-Z]{8}$`).MatchString(key)
		}), mock.MatchedBy(func(approval *sshEnrollment) bool {
			return approval.SessionUID == "session1" &&
				approval.SSHID == "root@namespace.device" &&
				approval.TenantID == "tenant1" &&
				approval.Username == "root" &&
				approval.IPAddress == "10.0.0.1" &&
				approval.RequestedAt.Equal(now) &&
				approval.State == models.SSHEnrollmentPending
		}), sshEnrollmentTTL).
		Return(nil).
		Once()

	service := NewService(storeMock, privateKey, publicKey, cacheMock, clientMock)

	approval, err := service.CreateSSHEnrollment(context.TODO(), req)
	require.NoError(t, err)
	require.Regexp(t, `^[2-9A-HJKMNP-TV-Z]{8}$`, approval.Code)
	require.Equal(t, int(sshEnrollmentTTL.Seconds()), approval.ExpiresIn)

	cacheMock.AssertExpectations(t)
	storeMock.AssertExpectations(t)
}

func TestGetSSHEnrollmentStatus(t *testing.T) {
	type Expected struct {
		status *models.SSHEnrollmentStatus
		err    error
	}

	cases := []struct {
		description   string
		code          string
		requiredMocks func(cacheMock *cachemock.MockCache)
		expected      Expected
	}{
		{
			description: "fails when the code is unknown or expired",
			code:        "WXYZ2K7Q",
			requiredMocks: func(cacheMock *cachemock.MockCache) {
				cacheMock.
					On("Get", mock.Anything, "ssh_enrollment/WXYZ2K7Q", mock.Anything).
					Return(nil).
					Once()
			},
			expected: Expected{status: nil, err: NewErrSSHEnrollmentCodeNotFound("WXYZ2K7Q", nil)},
		},
		{
			description: "reports pending while awaiting a decision",
			code:        "WXYZ2K7Q",
			requiredMocks: func(cacheMock *cachemock.MockCache) {
				cacheMock.
					On("Get", mock.Anything, "ssh_enrollment/WXYZ2K7Q", mock.Anything).
					Run(func(args mock.Arguments) {
						*args.Get(2).(*sshEnrollment) = sshEnrollment{State: models.SSHEnrollmentPending}
					}).
					Return(nil).
					Once()
			},
			expected: Expected{status: &models.SSHEnrollmentStatus{State: models.SSHEnrollmentPending}, err: nil},
		},
		{
			description: "returns the approver once approved",
			code:        "WXYZ2K7Q",
			requiredMocks: func(cacheMock *cachemock.MockCache) {
				cacheMock.
					On("Get", mock.Anything, "ssh_enrollment/WXYZ2K7Q", mock.Anything).
					Run(func(args mock.Arguments) {
						*args.Get(2).(*sshEnrollment) = sshEnrollment{State: models.SSHEnrollmentConfirmed, UserID: "owner1"}
					}).
					Return(nil).
					Once()
			},
			expected: Expected{status: &models.SSHEnrollmentStatus{State: models.SSHEnrollmentConfirmed, UserID: "owner1"}, err: nil},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			cacheMock := new(cachemock.MockCache)
			storeMock := new(storemock.MockStore)
			tc.requiredMocks(cacheMock)

			service := NewService(storeMock, privateKey, publicKey, cacheMock, clientMock)

			status, err := service.GetSSHEnrollmentStatus(context.TODO(), tc.code)
			require.Equal(tt, tc.expected.err, err)
			require.Equal(tt, tc.expected.status, status)

			cacheMock.AssertExpectations(tt)
		})
	}
}

func TestConfirmSSHEnrollment(t *testing.T) {
	namespace := &models.Namespace{
		Name:     "namespace1",
		TenantID: "tenant1",
		Members: []models.Member{
			{ID: "owner1", Role: authorizer.RoleOwner},
			{ID: "observer1", Role: authorizer.RoleObserver},
		},
	}

	populatePending := func(cacheMock *cachemock.MockCache) {
		cacheMock.
			On("Get", mock.Anything, "ssh_enrollment/WXYZ2K7Q", mock.Anything).
			Run(func(args mock.Arguments) {
				*args.Get(2).(*sshEnrollment) = sshEnrollment{
					SessionUID: "session1",
					TenantID:   "tenant1",
					State:      models.SSHEnrollmentPending,
				}
			}).
			Return(nil).
			Once()
	}

	cases := []struct {
		description   string
		userID        string
		code          string
		requiredMocks func(cacheMock *cachemock.MockCache, storeMock *storemock.MockStore)
		expectedErr   error
	}{
		{
			description:   "fails when the code is not well-formed",
			userID:        "owner1",
			code:          "00000000",
			requiredMocks: func(_ *cachemock.MockCache, _ *storemock.MockStore) {},
			expectedErr:   NewErrSSHEnrollmentCodeNotFound("00000000", nil),
		},
		{
			description: "fails when the code is unknown or expired",
			userID:      "owner1",
			code:        "WXYZ2K7Q",
			requiredMocks: func(cacheMock *cachemock.MockCache, _ *storemock.MockStore) {
				cacheMock.
					On("Get", mock.Anything, "ssh_enrollment/WXYZ2K7Q", mock.Anything).
					Return(nil).
					Once()
			},
			expectedErr: NewErrSSHEnrollmentCodeNotFound("WXYZ2K7Q", nil),
		},
		{
			description: "fails when the target namespace does not exist",
			userID:      "owner1",
			code:        "WXYZ2K7Q",
			requiredMocks: func(cacheMock *cachemock.MockCache, storeMock *storemock.MockStore) {
				populatePending(cacheMock)
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
			requiredMocks: func(cacheMock *cachemock.MockCache, storeMock *storemock.MockStore) {
				populatePending(cacheMock)
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
			requiredMocks: func(cacheMock *cachemock.MockCache, storeMock *storemock.MockStore) {
				populatePending(cacheMock)
				storeMock.
					On("NamespaceResolve", mock.Anything, store.NamespaceTenantIDResolver, "tenant1").
					Return(namespace, nil).
					Once()
			},
			expectedErr: NewErrRoleForbidden(),
		},
		{
			description: "approves and binds the approver",
			userID:      "owner1",
			code:        "WXYZ2K7Q",
			requiredMocks: func(cacheMock *cachemock.MockCache, storeMock *storemock.MockStore) {
				populatePending(cacheMock)
				storeMock.
					On("NamespaceResolve", mock.Anything, store.NamespaceTenantIDResolver, "tenant1").
					Return(namespace, nil).
					Once()
				cacheMock.
					On("SetNX", mock.Anything, "ssh_enrollment_decision/WXYZ2K7Q", "owner1", sshEnrollmentTTL).
					Return(true, nil).
					Once()
				cacheMock.
					On("Set", mock.Anything, "ssh_enrollment/WXYZ2K7Q", mock.MatchedBy(func(approval *sshEnrollment) bool {
						return approval.State == models.SSHEnrollmentConfirmed && approval.UserID == "owner1"
					}), sshEnrollmentTTL).
					Return(nil).
					Once()
			},
			expectedErr: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			cacheMock := new(cachemock.MockCache)
			storeMock := new(storemock.MockStore)
			tc.requiredMocks(cacheMock, storeMock)

			service := NewService(storeMock, privateKey, publicKey, cacheMock, clientMock)

			err := service.ConfirmSSHEnrollment(context.TODO(), tc.userID, &requests.SSHEnrollmentConfirm{Code: tc.code})
			require.Equal(tt, tc.expectedErr, err)

			cacheMock.AssertExpectations(tt)
			storeMock.AssertExpectations(tt)
		})
	}
}

func TestConfirmSSHEnrollmentEnrollment(t *testing.T) {
	namespace := &models.Namespace{
		Name:     "namespace1",
		TenantID: "00000000-0000-4000-0000-000000000000",
		Members:  []models.Member{{ID: "owner1", Role: authorizer.RoleOwner}},
	}

	cacheMock := new(cachemock.MockCache)
	storeMock := new(storemock.MockStore)
	queryOptionsMock := new(storemock.MockQueryOptions)
	storeMock.On("Options").Return(queryOptionsMock).Maybe()

	// An enrollment approval carries the presented key and the enroll flag.
	cacheMock.
		On("Get", mock.Anything, "ssh_enrollment/WXYZ2K7Q", mock.Anything).
		Run(func(args mock.Arguments) {
			*args.Get(2).(*sshEnrollment) = sshEnrollment{
				SessionUID:  "session1",
				TenantID:    namespace.TenantID,
				Fingerprint: "SHA256:abc",
				Data:        []byte("ssh-ed25519 AAAA"),
				Enroll:      true,
				State:       models.SSHEnrollmentPending,
			}
		}).
		Return(nil).
		Once()
	storeMock.
		On("NamespaceResolve", mock.Anything, store.NamespaceTenantIDResolver, namespace.TenantID).
		Return(namespace, nil).
		Once()
	cacheMock.
		On("SetNX", mock.Anything, "ssh_enrollment_decision/WXYZ2K7Q", "owner1", sshEnrollmentTTL).
		Return(true, nil).
		Once()

	// Accepting an enrollment binds the presented key to the approving account.
	queryOptionsMock.On("InNamespace", namespace.TenantID).Return(nil).Once()
	storeMock.
		On("SSHIdentityResolve", mock.Anything, store.SSHIdentityFingerprintResolver, "SHA256:abc", mock.Anything).
		Return(nil, store.ErrNoDocuments).
		Once()
	storeMock.
		On("SSHIdentityCreate", mock.Anything, mock.MatchedBy(func(identity *models.SSHIdentity) bool {
			return identity.PrincipalID == "owner1" && identity.Fingerprint == "SHA256:abc"
		})).
		Return("id1", nil).
		Once()
	cacheMock.
		On("Set", mock.Anything, "ssh_enrollment/WXYZ2K7Q", mock.MatchedBy(func(approval *sshEnrollment) bool {
			return approval.State == models.SSHEnrollmentConfirmed && approval.UserID == "owner1"
		}), sshEnrollmentTTL).
		Return(nil).
		Once()

	service := NewService(storeMock, privateKey, publicKey, cacheMock, clientMock)

	err := service.ConfirmSSHEnrollment(context.TODO(), "owner1", &requests.SSHEnrollmentConfirm{Code: "WXYZ2K7Q"})
	require.NoError(t, err)

	cacheMock.AssertExpectations(t)
	storeMock.AssertExpectations(t)
}

func TestAttachSSHEnrollmentKey(t *testing.T) {
	cacheMock := new(cachemock.MockCache)
	storeMock := new(storemock.MockStore)

	cacheMock.
		On("Get", mock.Anything, "ssh_enrollment/WXYZ2K7Q", mock.Anything).
		Run(func(args mock.Arguments) {
			*args.Get(2).(*sshEnrollment) = sshEnrollment{State: models.SSHEnrollmentPending}
		}).
		Return(nil).
		Once()
	cacheMock.
		On("Set", mock.Anything, "ssh_enrollment/WXYZ2K7Q", mock.MatchedBy(func(approval *sshEnrollment) bool {
			return approval.Fingerprint == "SHA256:abc" && approval.Enroll
		}), sshEnrollmentTTL).
		Return(nil).
		Once()

	service := NewService(storeMock, privateKey, publicKey, cacheMock, clientMock)

	err := service.AttachSSHEnrollmentKey(context.TODO(), &requests.SSHEnrollmentKey{
		Code:        "WXYZ2K7Q",
		Fingerprint: "SHA256:abc",
		Data:        []byte("ssh-ed25519 AAAA"),
		Enroll:      true,
	})
	require.NoError(t, err)

	cacheMock.AssertExpectations(t)
}
