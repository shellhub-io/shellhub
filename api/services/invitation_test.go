package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/api/pkg/responses"
	"github.com/shellhub-io/shellhub/api/store"
	storemock "github.com/shellhub-io/shellhub/api/store/mocks"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	storecache "github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/clock"
	clockmock "github.com/shellhub-io/shellhub/pkg/clock/mocks"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/pairingcode"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	uuidmock "github.com/shellhub-io/shellhub/pkg/uuid/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// mockClockNow swaps clock.DefaultBackend (and uuid.DefaultBackend) for mocks and restores them
// after the test. Invitation/register methods stamp CreatedAt/AddedAt and check expiry via
// clock.Now(); the token path mints a uuid. Isolating both keeps these tests from tripping over
// global mock state left by other tests in the package.
func mockClockNow(t *testing.T, now time.Time) {
	t.Helper()

	clockMock := clockmock.NewMockClock(t)
	prevClock := clock.DefaultBackend
	t.Cleanup(func() { clock.DefaultBackend = prevClock })
	clock.DefaultBackend = clockMock
	// Maybe(): early-return failure paths never reach a clock.Now() call.
	clockMock.On("Now").Return(now).Maybe()

	uuidMock := uuidmock.NewMockUUID(t)
	prevUUID := uuid.DefaultBackend
	t.Cleanup(func() { uuid.DefaultBackend = prevUUID })
	uuid.DefaultBackend = uuidMock
	uuidMock.On("Generate").Return("00000000-0000-4000-0000-000000000000").Maybe()
}

func TestService_ResolveInvitation(t *testing.T) {
	storeMock := storemock.NewMockStore(t)
	ctx := context.TODO()

	// A freshly-minted code passes pairingcode.IsValid; the service normalizes before lookup.
	code, err := pairingcode.New(pairingcode.InviteCodeLength)
	assert.NoError(t, err)
	normalized := pairingcode.Normalize(code)

	type Expected struct {
		resp *responses.ResolveInvitation
		err  error
	}

	cases := []struct {
		description   string
		req           *requests.ResolveInvitation
		requiredMocks func()
		expected      Expected
	}{
		{
			description:   "fails when the code is not a valid pairing code",
			req:           &requests.ResolveInvitation{Invite: "not-a-code"},
			requiredMocks: func() {},
			expected:      Expected{nil, NewErrAuthForbidden()},
		},
		{
			description: "fails when no invitation carries the signature",
			req:         &requests.ResolveInvitation{Invite: code},
			requiredMocks: func() {
				storeMock.On("MembershipInvitationResolveBySig", ctx, normalized).
					Return(nil, store.ErrNoDocuments).Once()
			},
			expected: Expected{nil, NewErrAuthForbidden()},
		},
		{
			description: "resolves to the invited placeholder when no real account exists yet",
			req:         &requests.ResolveInvitation{Invite: code},
			requiredMocks: func() {
				storeMock.On("MembershipInvitationResolveBySig", ctx, normalized).
					Return(&models.MembershipInvitation{TenantID: "tenant", UserID: "user"}, nil).Once()
				storeMock.On("UserResolve", ctx, store.UserIDResolver, "user").
					Return(nil, store.ErrNoDocuments).Once()
				storeMock.On("UserInvitationGet", ctx, store.UserInvitationIDResolver, "user").
					Return(&models.UserInvitation{ID: "user", Email: "invitee@test.com"}, nil).Once()
			},
			expected: Expected{
				&responses.ResolveInvitation{TenantID: "tenant", UserID: "user", Email: "invitee@test.com", Status: "invited"},
				nil,
			},
		},
		{
			description: "resolves to the real account status when it already exists",
			req:         &requests.ResolveInvitation{Invite: code},
			requiredMocks: func() {
				storeMock.On("MembershipInvitationResolveBySig", ctx, normalized).
					Return(&models.MembershipInvitation{TenantID: "tenant", UserID: "user"}, nil).Once()
				storeMock.On("UserResolve", ctx, store.UserIDResolver, "user").
					Return(&models.User{
						ID:       "user",
						Status:   models.UserStatusConfirmed,
						UserData: models.UserData{Email: "real@test.com"},
					}, nil).Once()
			},
			expected: Expected{
				&responses.ResolveInvitation{TenantID: "tenant", UserID: "user", Email: "real@test.com", Status: "confirmed"},
				nil,
			},
		},
	}

	s := NewService(store.Store(storeMock), privateKey, publicKey, storecache.NewNullCache(), clientMock)

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			resp, err := s.ResolveInvitation(ctx, tc.req)
			assert.Equal(t, tc.expected.resp, resp)
			assert.Equal(t, tc.expected.err, err)
		})
	}

	storeMock.AssertExpectations(t)
}

func TestService_AcceptInvite(t *testing.T) {
	storeMock := storemock.NewMockStore(t)
	ctx := context.TODO()

	now := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	future := now.Add(time.Hour)
	past := now.Add(-time.Hour)

	req := &requests.AcceptInvite{TenantID: "tenant", UserID: "user"}

	cases := []struct {
		description   string
		requiredMocks func()
		expected      error
	}{
		{
			description: "fails when the user does not exist",
			requiredMocks: func() {
				storeMock.On("UserResolve", ctx, store.UserIDResolver, "user").
					Return(nil, store.ErrNoDocuments).Once()
			},
			expected: NewErrUserNotFound("user", store.ErrNoDocuments),
		},
		{
			description: "fails when the namespace does not exist",
			requiredMocks: func() {
				storeMock.On("UserResolve", ctx, store.UserIDResolver, "user").
					Return(&models.User{ID: "user"}, nil).Once()
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "tenant").
					Return(nil, store.ErrNoDocuments).Once()
			},
			expected: NewErrNamespaceNotFound("tenant", store.ErrNoDocuments),
		},
		{
			description: "fails when the user is already a member",
			requiredMocks: func() {
				storeMock.On("UserResolve", ctx, store.UserIDResolver, "user").
					Return(&models.User{ID: "user"}, nil).Once()
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "tenant").
					Return(&models.Namespace{
						TenantID: "tenant",
						Members:  []models.Member{{ID: "user", Role: authorizer.RoleObserver}},
					}, nil).Once()
			},
			expected: NewErrNamespaceMemberDuplicated("user", nil),
		},
		{
			description: "fails when the invitation is not pending",
			requiredMocks: func() {
				storeMock.On("UserResolve", ctx, store.UserIDResolver, "user").
					Return(&models.User{ID: "user"}, nil).Once()
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "tenant").
					Return(&models.Namespace{TenantID: "tenant"}, nil).Once()
				storeMock.On("MembershipInvitationResolve", ctx, "tenant", "user").
					Return(&models.MembershipInvitation{
						Status:    models.MembershipInvitationStatusCancelled,
						ExpiresAt: &future,
					}, nil).Once()
			},
			expected: NewErrNamespaceMemberNotFound("user", nil),
		},
		{
			description: "fails when the invitation is expired",
			requiredMocks: func() {
				storeMock.On("UserResolve", ctx, store.UserIDResolver, "user").
					Return(&models.User{ID: "user"}, nil).Once()
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "tenant").
					Return(&models.Namespace{TenantID: "tenant"}, nil).Once()
				storeMock.On("MembershipInvitationResolve", ctx, "tenant", "user").
					Return(&models.MembershipInvitation{
						Status:    models.MembershipInvitationStatusPending,
						ExpiresAt: &past,
					}, nil).Once()
			},
			expected: NewErrNamespaceMemberNotFound("user", nil),
		},
		{
			description: "succeeds joining the namespace and consuming the invitation",
			requiredMocks: func() {
				storeMock.On("UserResolve", ctx, store.UserIDResolver, "user").
					Return(&models.User{ID: "user"}, nil).Once()
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "tenant").
					Return(&models.Namespace{TenantID: "tenant"}, nil).Once()
				invitation := &models.MembershipInvitation{
					TenantID:  "tenant",
					UserID:    "user",
					Status:    models.MembershipInvitationStatusPending,
					Role:      authorizer.RoleOperator,
					ExpiresAt: &future,
				}
				storeMock.On("MembershipInvitationResolve", ctx, "tenant", "user").
					Return(invitation, nil).Once()
				storeMock.On("WithTransaction", ctx, mock.AnythingOfType("store.TransactionCb")).
					Return(func(_ context.Context, cb store.TransactionCb) error { return cb(ctx) }).Once()
				storeMock.On("NamespaceCreateMembership", ctx, "tenant", &models.Member{
					ID:      "user",
					AddedAt: now,
					Role:    authorizer.RoleOperator,
				}).Return(nil).Once()
				storeMock.On("MembershipInvitationDelete", ctx, invitation).Return(nil).Once()
			},
			expected: nil,
		},
	}

	s := NewService(store.Store(storeMock), privateKey, publicKey, storecache.NewNullCache(), clientMock)

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			mockClockNow(t, now)
			tc.requiredMocks()

			err := s.AcceptInvite(ctx, req)
			assert.Equal(t, tc.expected, err)
		})
	}

	storeMock.AssertExpectations(t)
}

func TestService_GenerateInvitationLink(t *testing.T) {
	storeMock := storemock.NewMockStore(t)
	ctx := context.TODO()

	now := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)

	owner := models.Member{ID: "owner", Role: authorizer.RoleOwner}
	namespace := &models.Namespace{TenantID: "tenant", Members: []models.Member{owner}}

	req := &requests.GenerateInvitationLink{
		ForwardedHost: "shellhub.test",
		TenantID:      "tenant",
		UserID:        "owner",
		MemberEmail:   "invitee@test.com",
		MemberRole:    authorizer.RoleOperator,
	}

	type Expected struct {
		hasLink bool
		err     error
	}

	cases := []struct {
		description      string
		directMembership bool
		requiredMocks    func()
		expected         Expected
	}{
		{
			description: "fails when the namespace does not exist",
			requiredMocks: func() {
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "tenant").
					Return(nil, store.ErrNoDocuments).Once()
			},
			expected: Expected{false, NewErrNamespaceNotFound("tenant", store.ErrNoDocuments)},
		},
		{
			description: "fails when the active member is not on the namespace",
			requiredMocks: func() {
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "tenant").
					Return(&models.Namespace{TenantID: "tenant"}, nil).Once()
			},
			expected: Expected{false, NewErrNamespaceMemberNotFound("owner", nil)},
		},
		{
			description: "fails when the active member cannot act over the invited role",
			requiredMocks: func() {
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "tenant").
					Return(&models.Namespace{
						TenantID: "tenant",
						Members:  []models.Member{{ID: "owner", Role: authorizer.RoleObserver}},
					}, nil).Once()
			},
			expected: Expected{false, NewErrRoleForbidden()},
		},
		{
			description: "fails when the invited account is already a member",
			requiredMocks: func() {
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "tenant").
					Return(&models.Namespace{
						TenantID: "tenant",
						Members:  []models.Member{owner, {ID: "invitee", Role: authorizer.RoleObserver}},
					}, nil).Once()
				storeMock.On("WithTransaction", ctx, mock.AnythingOfType("store.TransactionCb")).
					Return(func(ctx context.Context, cb store.TransactionCb) error { return cb(ctx) }).Once()
				storeMock.On("UserResolve", ctx, store.UserEmailResolver, "invitee@test.com").
					Return(&models.User{ID: "invitee"}, nil).Once()
			},
			expected: Expected{false, NewErrNamespaceMemberDuplicated("invitee", nil)},
		},
		{
			description: "creates the invitation for an existing account and returns the link",
			requiredMocks: func() {
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "tenant").
					Return(namespace, nil).Once()
				storeMock.On("WithTransaction", ctx, mock.AnythingOfType("store.TransactionCb")).
					Return(func(ctx context.Context, cb store.TransactionCb) error { return cb(ctx) }).Once()
				storeMock.On("UserResolve", ctx, store.UserEmailResolver, "invitee@test.com").
					Return(&models.User{ID: "invitee"}, nil).Once()
				storeMock.On("MembershipInvitationResolve", ctx, "tenant", "invitee").
					Return(nil, store.ErrNoDocuments).Once()
				storeMock.On("MembershipInvitationCreate", ctx, mock.AnythingOfType("*models.MembershipInvitation")).
					Return(nil).Once()
			},
			expected: Expected{true, nil},
		},
		{
			description: "upserts a placeholder invitation when no account exists yet",
			requiredMocks: func() {
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "tenant").
					Return(namespace, nil).Once()
				storeMock.On("WithTransaction", ctx, mock.AnythingOfType("store.TransactionCb")).
					Return(func(ctx context.Context, cb store.TransactionCb) error { return cb(ctx) }).Once()
				storeMock.On("UserResolve", ctx, store.UserEmailResolver, "invitee@test.com").
					Return(nil, store.ErrNoDocuments).Once()
				storeMock.On("UserInvitationsUpsert", ctx, "invitee@test.com").
					Return("placeholder", nil).Once()
				storeMock.On("MembershipInvitationResolve", ctx, "tenant", "placeholder").
					Return(nil, store.ErrNoDocuments).Once()
				storeMock.On("MembershipInvitationCreate", ctx, mock.AnythingOfType("*models.MembershipInvitation")).
					Return(nil).Once()
			},
			expected: Expected{true, nil},
		},
		{
			description: "resends an expired invitation and returns a fresh link",
			requiredMocks: func() {
				expired := now.Add(-time.Hour)
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "tenant").
					Return(namespace, nil).Once()
				storeMock.On("WithTransaction", ctx, mock.AnythingOfType("store.TransactionCb")).
					Return(func(ctx context.Context, cb store.TransactionCb) error { return cb(ctx) }).Once()
				storeMock.On("UserResolve", ctx, store.UserEmailResolver, "invitee@test.com").
					Return(&models.User{ID: "invitee"}, nil).Once()
				storeMock.On("MembershipInvitationResolve", ctx, "tenant", "invitee").
					Return(&models.MembershipInvitation{
						TenantID:  "tenant",
						UserID:    "invitee",
						Status:    models.MembershipInvitationStatusPending,
						ExpiresAt: &expired,
					}, nil).Once()
				storeMock.On("MembershipInvitationUpdate", ctx, mock.AnythingOfType("*models.MembershipInvitation")).
					Return(nil).Once()
			},
			expected: Expected{true, nil},
		},
		{
			description:      "adds an existing account directly and returns no link when direct membership is on (enterprise)",
			directMembership: true,
			requiredMocks: func() {
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "tenant").
					Return(namespace, nil).Once()
				storeMock.On("WithTransaction", ctx, mock.AnythingOfType("store.TransactionCb")).
					Return(func(ctx context.Context, cb store.TransactionCb) error { return cb(ctx) }).Once()
				storeMock.On("UserResolve", ctx, store.UserEmailResolver, "invitee@test.com").
					Return(&models.User{ID: "invitee"}, nil).Once()
				storeMock.On("NamespaceCreateMembership", ctx, "tenant", &models.Member{
					ID: "invitee", AddedAt: now, Role: authorizer.RoleOperator,
				}).Return(nil).Once()
			},
			expected: Expected{false, nil},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			directMembershipEnabled = tc.directMembership
			t.Cleanup(func() { directMembershipEnabled = false })

			mockClockNow(t, now)
			tc.requiredMocks()

			s := NewService(store.Store(storeMock), privateKey, publicKey, storecache.NewNullCache(), clientMock)

			link, err := s.GenerateInvitationLink(ctx, req)
			assert.Equal(t, tc.expected.err, err)
			if tc.expected.hasLink {
				assert.Contains(t, link, "shellhub.test/accept-invite?invite=")
			} else {
				assert.Empty(t, link)
			}
		})
	}

	storeMock.AssertExpectations(t)
}

func TestService_UserMembershipInvitationList(t *testing.T) {
	storeMock := storemock.NewMockStore(t)
	queryOptionsMock := storemock.NewMockQueryOptions(t)
	storeMock.On("Options").Return(queryOptionsMock)
	ctx := context.TODO()

	req := &requests.UserMembershipInvitationList{
		UserID:    "user",
		Paginator: query.Paginator{Page: 1, PerPage: 10},
	}

	type Expected struct {
		invitations []responses.MembershipInvitation
		count       int64
		err         error
	}

	cases := []struct {
		description   string
		requiredMocks func()
		expected      Expected
	}{
		{
			description: "fails when the store returns an error",
			requiredMocks: func() {
				queryOptionsMock.On("Match", &req.Filters).Return(nil).Once()
				queryOptionsMock.On("Sort", &req.Sorter).Return(nil).Once()
				queryOptionsMock.On("Paginate", &req.Paginator).Return(nil).Once()
				storeMock.On("UserMembershipInvitationList", ctx, "user", mock.AnythingOfType("[]store.QueryOption")).
					Return(nil, int64(0), errors.New("error")).Once()
			},
			expected: Expected{nil, 0, errors.New("error")},
		},
		{
			description: "succeeds listing the user's invitations",
			requiredMocks: func() {
				queryOptionsMock.On("Match", &req.Filters).Return(nil).Once()
				queryOptionsMock.On("Sort", &req.Sorter).Return(nil).Once()
				queryOptionsMock.On("Paginate", &req.Paginator).Return(nil).Once()
				storeMock.On("UserMembershipInvitationList", ctx, "user", mock.AnythingOfType("[]store.QueryOption")).
					Return([]models.MembershipInvitation{
						{TenantID: "tenant", UserID: "user", Role: authorizer.RoleOperator},
					}, int64(1), nil).Once()
			},
			expected: Expected{
				invitations: []responses.MembershipInvitation{
					*responses.MembershipInvitationFromModel(&models.MembershipInvitation{
						TenantID: "tenant", UserID: "user", Role: authorizer.RoleOperator,
					}),
				},
				count: 1,
				err:   nil,
			},
		},
	}

	s := NewService(store.Store(storeMock), privateKey, publicKey, storecache.NewNullCache(), clientMock)

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			invitations, count, err := s.UserMembershipInvitationList(ctx, req)
			assert.Equal(t, tc.expected.invitations, invitations)
			assert.Equal(t, tc.expected.count, count)
			assert.Equal(t, tc.expected.err, err)
		})
	}

	storeMock.AssertExpectations(t)
}

func TestService_NamespaceMembershipInvitationList(t *testing.T) {
	storeMock := storemock.NewMockStore(t)
	queryOptionsMock := storemock.NewMockQueryOptions(t)
	ctx := context.TODO()

	req := &requests.NamespaceMembershipInvitationList{
		TenantID:  "tenant",
		UserID:    "owner",
		Paginator: query.Paginator{Page: 1, PerPage: 10},
	}

	adminNamespace := &models.Namespace{
		TenantID: "tenant",
		Members:  []models.Member{{ID: "owner", Role: authorizer.RoleOwner}},
	}

	type Expected struct {
		invitations []responses.MembershipInvitation
		count       int64
		err         error
	}

	cases := []struct {
		description   string
		requiredMocks func()
		expected      Expected
	}{
		{
			description: "fails when the namespace does not exist",
			requiredMocks: func() {
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "tenant").
					Return(nil, store.ErrNoDocuments).Once()
			},
			expected: Expected{nil, 0, NewErrNamespaceNotFound("tenant", store.ErrNoDocuments)},
		},
		{
			description: "fails when the active user is not a member",
			requiredMocks: func() {
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "tenant").
					Return(&models.Namespace{TenantID: "tenant"}, nil).Once()
			},
			expected: Expected{nil, 0, NewErrNamespaceMemberNotFound("owner", nil)},
		},
		{
			description: "fails when the active user's role is below administrator",
			requiredMocks: func() {
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "tenant").
					Return(&models.Namespace{
						TenantID: "tenant",
						Members:  []models.Member{{ID: "owner", Role: authorizer.RoleObserver}},
					}, nil).Once()
			},
			expected: Expected{nil, 0, NewErrRoleForbidden()},
		},
		{
			description: "fails when the store returns an error",
			requiredMocks: func() {
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "tenant").
					Return(adminNamespace, nil).Once()
				storeMock.On("Options").Return(queryOptionsMock).Once()
				queryOptionsMock.On("Match", &req.Filters).Return(nil).Once()
				storeMock.On("Options").Return(queryOptionsMock).Once()
				queryOptionsMock.On("Sort", &req.Sorter).Return(nil).Once()
				storeMock.On("Options").Return(queryOptionsMock).Once()
				queryOptionsMock.On("Paginate", &req.Paginator).Return(nil).Once()
				storeMock.On("NamespaceMembershipInvitationList", ctx, "tenant", mock.AnythingOfType("[]store.QueryOption")).
					Return(nil, int64(0), errors.New("error")).Once()
			},
			expected: Expected{nil, 0, errors.New("error")},
		},
		{
			description: "succeeds listing the namespace's invitations",
			requiredMocks: func() {
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "tenant").
					Return(adminNamespace, nil).Once()
				storeMock.On("Options").Return(queryOptionsMock).Once()
				queryOptionsMock.On("Match", &req.Filters).Return(nil).Once()
				storeMock.On("Options").Return(queryOptionsMock).Once()
				queryOptionsMock.On("Sort", &req.Sorter).Return(nil).Once()
				storeMock.On("Options").Return(queryOptionsMock).Once()
				queryOptionsMock.On("Paginate", &req.Paginator).Return(nil).Once()
				storeMock.On("NamespaceMembershipInvitationList", ctx, "tenant", mock.AnythingOfType("[]store.QueryOption")).
					Return([]models.MembershipInvitation{
						{TenantID: "tenant", UserID: "invitee", Role: authorizer.RoleOperator},
					}, int64(1), nil).Once()
			},
			expected: Expected{
				invitations: []responses.MembershipInvitation{
					*responses.MembershipInvitationFromModel(&models.MembershipInvitation{
						TenantID: "tenant", UserID: "invitee", Role: authorizer.RoleOperator,
					}),
				},
				count: 1,
				err:   nil,
			},
		},
	}

	s := NewService(store.Store(storeMock), privateKey, publicKey, storecache.NewNullCache(), clientMock)

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			invitations, count, err := s.NamespaceMembershipInvitationList(ctx, req)
			assert.Equal(t, tc.expected.invitations, invitations)
			assert.Equal(t, tc.expected.count, count)
			assert.Equal(t, tc.expected.err, err)
		})
	}

	storeMock.AssertExpectations(t)
}

func TestService_CancelMembershipInvitation(t *testing.T) {
	storeMock := storemock.NewMockStore(t)
	ctx := context.TODO()

	now := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)

	req := &requests.CancelMembershipInvitation{
		TenantID:      "tenant",
		UserID:        "owner",
		InvitedUserID: "invitee",
	}

	adminNamespace := &models.Namespace{
		TenantID: "tenant",
		Members:  []models.Member{{ID: "owner", Role: authorizer.RoleOwner}},
	}

	cases := []struct {
		description   string
		requiredMocks func()
		expected      error
	}{
		{
			description: "fails when the namespace does not exist",
			requiredMocks: func() {
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "tenant").
					Return(nil, store.ErrNoDocuments).Once()
			},
			expected: NewErrNamespaceNotFound("tenant", store.ErrNoDocuments),
		},
		{
			description: "fails when the active user is not a member",
			requiredMocks: func() {
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "tenant").
					Return(&models.Namespace{TenantID: "tenant"}, nil).Once()
			},
			expected: NewErrNamespaceMemberNotFound("owner", nil),
		},
		{
			description: "fails when the invitation does not exist",
			requiredMocks: func() {
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "tenant").
					Return(adminNamespace, nil).Once()
				storeMock.On("MembershipInvitationResolve", ctx, "tenant", "invitee").
					Return(nil, store.ErrNoDocuments).Once()
			},
			expected: NewErrNamespaceMemberNotFound("invitee", store.ErrNoDocuments),
		},
		{
			description: "fails when the invitation is not pending",
			requiredMocks: func() {
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "tenant").
					Return(adminNamespace, nil).Once()
				storeMock.On("MembershipInvitationResolve", ctx, "tenant", "invitee").
					Return(&models.MembershipInvitation{Status: models.MembershipInvitationStatusCancelled}, nil).Once()
			},
			expected: NewErrNamespaceMemberNotFound("invitee", nil),
		},
		{
			description: "fails when the active user has no authority over the invitation role",
			requiredMocks: func() {
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "tenant").
					Return(&models.Namespace{
						TenantID: "tenant",
						Members:  []models.Member{{ID: "owner", Role: authorizer.RoleOperator}},
					}, nil).Once()
				storeMock.On("MembershipInvitationResolve", ctx, "tenant", "invitee").
					Return(&models.MembershipInvitation{
						Status: models.MembershipInvitationStatusPending,
						Role:   authorizer.RoleAdministrator,
					}, nil).Once()
			},
			expected: NewErrRoleForbidden(),
		},
		{
			description: "succeeds cancelling the invitation",
			requiredMocks: func() {
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "tenant").
					Return(adminNamespace, nil).Once()
				storeMock.On("MembershipInvitationResolve", ctx, "tenant", "invitee").
					Return(&models.MembershipInvitation{
						Status: models.MembershipInvitationStatusPending,
						Role:   authorizer.RoleOperator,
					}, nil).Once()
				storeMock.On("MembershipInvitationUpdate", ctx, mock.MatchedBy(func(m *models.MembershipInvitation) bool {
					return m.Status == models.MembershipInvitationStatusCancelled && m.StatusUpdatedAt.Equal(now)
				})).Return(nil).Once()
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			mockClockNow(t, now)
			tc.requiredMocks()

			s := NewService(store.Store(storeMock), privateKey, publicKey, storecache.NewNullCache(), clientMock)

			err := s.CancelMembershipInvitation(ctx, req)
			assert.Equal(t, tc.expected, err)
		})
	}

	storeMock.AssertExpectations(t)
}
