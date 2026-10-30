package routes

import (
	"encoding/json"
	stderrors "errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
	svc "github.com/shellhub-io/shellhub/server/api/services"
	"github.com/shellhub-io/shellhub/server/api/services/mocks"
	"github.com/stretchr/testify/assert"
	gomock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUpdateUser(t *testing.T) {
	type Expected struct {
		status int
	}

	svcMock := mocks.NewMockService(t)

	cases := []struct {
		description   string
		headers       map[string]string
		body          requests.UpdateUser
		requiredMocks func()
		expected      Expected
	}{
		{
			description: "fails when bind fails to validate e-mail",
			headers: map[string]string{
				"X-ID":   "000000000000000000000000",
				"X-Role": "owner",
			},
			body: requests.UpdateUser{
				Name:          "John Doe",
				Username:      "john_doe",
				Email:         "invalid.com",
				RecoveryEmail: "invalid.com",
			},
			requiredMocks: func() {},
			expected:      Expected{http.StatusBadRequest},
		},
		{
			description: "fails when bind fails to validate username",
			headers: map[string]string{
				"X-ID":   "000000000000000000000000",
				"X-Role": "owner",
			},
			body: requests.UpdateUser{
				Name:          "John Doe",
				Username:      "_",
				Email:         "john.doe@test.com",
				RecoveryEmail: "john.doe@test.com",
			},
			requiredMocks: func() {},
			expected:      Expected{http.StatusBadRequest},
		},
		{
			description: "fails when try to updating a non-existing user",
			headers: map[string]string{
				"X-ID":   "000000000000000000000000",
				"X-Role": "owner",
			},
			body: requests.UpdateUser{
				Name:          "John Doe",
				Username:      "john_doe",
				Email:         "john.doe@test.com",
				RecoveryEmail: "john.doe@test.com",
			},
			requiredMocks: func() {
				svcMock.
					On(
						"UpdateUser",
						gomock.Anything,
						&requests.UpdateUser{
							UserID:        "000000000000000000000000",
							Name:          "John Doe",
							Username:      "john_doe",
							Email:         "john.doe@test.com",
							RecoveryEmail: "john.doe@test.com",
						},
					).
					Return(svc.ErrUserNotFound).
					Once()
			},
			expected: Expected{http.StatusNotFound},
		},
		{
			description: "success when try to updating an existing user",
			body: requests.UpdateUser{
				Name:          "John Doe",
				Username:      "john_doe",
				Email:         "john.doe@test.com",
				RecoveryEmail: "john.doe@test.com",
			},
			headers: map[string]string{
				"X-ID":   "000000000000000000000000",
				"X-Role": "owner",
			},
			requiredMocks: func() {
				svcMock.
					On(
						"UpdateUser",
						gomock.Anything,
						&requests.UpdateUser{
							UserID:        "000000000000000000000000",
							Name:          "John Doe",
							Username:      "john_doe",
							Email:         "john.doe@test.com",
							RecoveryEmail: "john.doe@test.com",
						},
					).
					Return(nil).
					Once()
			},
			expected: Expected{http.StatusOK},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			data, err := json.Marshal(tc.body) //nolint:gosec // G117: test request serialization
			require.NoError(t, err)

			req := httptest.NewRequestWithContext(t.Context(), http.MethodPatch, "/api/users", strings.NewReader(string(data)))
			req.Header.Set("Content-Type", "application/json")
			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}

			rec := httptest.NewRecorder()

			e := newTestRouter(t, svcMock)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expected, Expected{rec.Result().StatusCode})
		})
	}

	svcMock.AssertExpectations(t)
}

func TestUpdateUserPassword(t *testing.T) {
	mock := mocks.NewMockService(t)

	cases := []struct {
		title             string
		updatePayloadMock requests.UserPasswordUpdate
		requiredMocks     func(updatePayloadMock requests.UserPasswordUpdate)
		expectedStatus    int
	}{
		{
			title: "fails when validate because the password fields are missing",
			updatePayloadMock: requests.UserPasswordUpdate{
				UserID: "123",
			},
			requiredMocks:  func(_ requests.UserPasswordUpdate) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			title: "fails when validate because the tag does not have a min of 5 characters",
			updatePayloadMock: requests.UserPasswordUpdate{
				UserID:          "123",
				CurrentPassword: "fail",
				NewPassword:     "new_password",
			},
			requiredMocks:  func(_ requests.UserPasswordUpdate) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			title: "fails when validate because the tag does not have a max of 32 characters",
			updatePayloadMock: requests.UserPasswordUpdate{ //nolint:gosec // G101: test fixture
				UserID:          "123",
				CurrentPassword: "1a3b8f0c2e5d7g9i4k6m8o2q5s7u9w1v7",
				NewPassword:     "new_password",
			},
			requiredMocks:  func(_ requests.UserPasswordUpdate) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			title: "fails when validate because the tag does not have a min of 5 characters",
			updatePayloadMock: requests.UserPasswordUpdate{
				UserID:          "123",
				CurrentPassword: "new_password",
				NewPassword:     "fail",
			},
			requiredMocks:  func(_ requests.UserPasswordUpdate) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			title: "fails when validate because the tag does not have a max of 32 characters",
			updatePayloadMock: requests.UserPasswordUpdate{ //nolint:gosec // G101: test fixture
				UserID:          "123",
				CurrentPassword: "new_password",
				NewPassword:     "1a3b8f0c2e5d7g9i4k6m8o2q5s7u9w1v7",
			},
			requiredMocks:  func(_ requests.UserPasswordUpdate) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			title: "fails when validate because have a duplicate password",
			updatePayloadMock: requests.UserPasswordUpdate{
				UserID:          "123",
				NewPassword:     "duplicate",
				CurrentPassword: "duplicate",
			},
			requiredMocks:  func(_ requests.UserPasswordUpdate) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			title: "fails when current password does not match",
			updatePayloadMock: requests.UserPasswordUpdate{
				UserID:          "123",
				CurrentPassword: "wrong_password",
				NewPassword:     "new_password",
			},
			requiredMocks: func(updatePayloadMock requests.UserPasswordUpdate) {
				mock.On("UpdatePasswordUser", gomock.Anything, "123", updatePayloadMock.CurrentPassword, updatePayloadMock.NewPassword).Return(svc.ErrUserPasswordNotMatch).Once()
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			title: "fails when try to updating a password an existing user",
			updatePayloadMock: requests.UserPasswordUpdate{
				UserID:          "123",
				CurrentPassword: "old_password",
				NewPassword:     "new_password",
			},
			requiredMocks: func(updatePayloadMock requests.UserPasswordUpdate) {
				mock.On("UpdatePasswordUser", gomock.Anything, "123", updatePayloadMock.CurrentPassword, updatePayloadMock.NewPassword).Return(svc.ErrUserNotFound).Once()
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			title: "success when try to updating a password an existing user",
			updatePayloadMock: requests.UserPasswordUpdate{
				UserID:          "123",
				CurrentPassword: "old_password",
				NewPassword:     "new_password",
			},
			requiredMocks: func(updatePayloadMock requests.UserPasswordUpdate) {
				mock.On("UpdatePasswordUser", gomock.Anything, "123", updatePayloadMock.CurrentPassword, updatePayloadMock.NewPassword).Return(nil).Once()
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range cases {
		t.Run(tc.title, func(t *testing.T) {
			tc.requiredMocks(tc.updatePayloadMock)

			jsonData, err := json.Marshal(tc.updatePayloadMock)
			if err != nil {
				require.NoError(t, err)
			}

			req := httptest.NewRequestWithContext(t.Context(), http.MethodPatch, "/api/users/path-id-ignored/password", strings.NewReader(string(jsonData)))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-ID", tc.updatePayloadMock.UserID)
			req.Header.Set("X-Role", authorizer.RoleOwner.String())
			rec := httptest.NewRecorder()

			e := newTestRouter(t, mock)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatus, rec.Result().StatusCode)
		})
	}

	mock.AssertExpectations(t)
}

func TestUpdateUserPasswordIgnoresPathID(t *testing.T) {
	mock := mocks.NewMockService(t)

	const sessionID = "session-user-id"
	const attackerTargetID = "victim-user-id"

	mock.On("UpdatePasswordUser", gomock.Anything, sessionID, "old_password", "new_password").
		Return(nil).Once()

	payload := requests.UserPasswordUpdate{
		CurrentPassword: "old_password",
		NewPassword:     "new_password",
	}

	jsonData, err := json.Marshal(payload)
	require.NoError(t, err)

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPatch, fmt.Sprintf("/api/users/%s/password", attackerTargetID), strings.NewReader(string(jsonData)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-ID", sessionID)
	req.Header.Set("X-Role", authorizer.RoleObserver.String())

	rec := httptest.NewRecorder()

	e := newTestRouter(t, mock)
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Result().StatusCode)
	mock.AssertExpectations(t)
}

func TestRevokeUserTokens(t *testing.T) {
	cases := []struct {
		description   string
		headers       map[string]string
		requiredMocks func(svcMock *mocks.MockService)
		expected      int
	}{
		{
			description:   "fails when the caller has no ID",
			headers:       map[string]string{"X-Role": authorizer.RoleObserver.String()},
			requiredMocks: func(_ *mocks.MockService) {},
			expected:      http.StatusBadRequest,
		},
		{
			description: "fails when the user does not exist",
			headers: map[string]string{
				"X-ID":   "000000000000000000000000",
				"X-Role": authorizer.RoleObserver.String(),
			},
			requiredMocks: func(svcMock *mocks.MockService) {
				svcMock.
					On("RevokeUserTokens", gomock.Anything, "000000000000000000000000").
					Return(svc.ErrUserNotFound).
					Once()
			},
			expected: http.StatusNotFound,
		},
		{
			description: "succeeds when the user exists",
			headers: map[string]string{
				"X-ID":   "000000000000000000000000",
				"X-Role": authorizer.RoleObserver.String(),
			},
			requiredMocks: func(svcMock *mocks.MockService) {
				svcMock.
					On("RevokeUserTokens", gomock.Anything, "000000000000000000000000").
					Return(nil).
					Once()
			},
			expected: http.StatusOK,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			svcMock := mocks.NewMockService(t)
			tc.requiredMocks(svcMock)

			req := httptest.NewRequestWithContext(t.Context(), http.MethodDelete, "/api/users/tokens", nil)
			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}

			rec := httptest.NewRecorder()

			e := newTestRouter(t, svcMock)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expected, rec.Result().StatusCode)
		})
	}
}

func TestRecordOnboardingSurvey(t *testing.T) {
	const userID = "000000000000000000000000"

	cases := []struct {
		description   string
		outcome       string
		requiredMocks func(svcMock *mocks.MockService)
		expected      int
	}{
		{
			description:   "refuses pending",
			outcome:       "pending",
			requiredMocks: func(_ *mocks.MockService) {},
			expected:      http.StatusBadRequest,
		},
		{
			description: "fails when the user does not exist",
			outcome:     "skipped",
			requiredMocks: func(svcMock *mocks.MockService) {
				svcMock.
					On("RecordOnboardingSurvey", gomock.Anything, &requests.RecordOnboardingSurvey{UserID: userID, Outcome: models.UserOnboardingSurveySkipped}).
					Return(svc.ErrUserNotFound).
					Once()
			},
			expected: http.StatusNotFound,
		},
		{
			description: "fails when the outcome cannot be saved",
			outcome:     "skipped",
			requiredMocks: func(svcMock *mocks.MockService) {
				svcMock.
					On("RecordOnboardingSurvey", gomock.Anything, &requests.RecordOnboardingSurvey{UserID: userID, Outcome: models.UserOnboardingSurveySkipped}).
					Return(svc.NewErrUserUpdate(&models.User{ID: userID}, stderrors.New("error"))).
					Once()
			},
			expected: http.StatusInternalServerError,
		},
		{
			description: "records an answer",
			outcome:     "answered",
			requiredMocks: func(svcMock *mocks.MockService) {
				svcMock.
					On("RecordOnboardingSurvey", gomock.Anything, &requests.RecordOnboardingSurvey{UserID: userID, Outcome: models.UserOnboardingSurveyAnswered}).
					Return(nil).
					Once()
			},
			expected: http.StatusOK,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			svcMock := mocks.NewMockService(t)
			tc.requiredMocks(svcMock)

			body := strings.NewReader(fmt.Sprintf(`{"outcome":%q}`, tc.outcome))
			req := httptest.NewRequestWithContext(t.Context(), http.MethodPut, "/api/users/onboarding-survey", body)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-ID", userID)
			req.Header.Set("X-Role", authorizer.RoleObserver.String())

			rec := httptest.NewRecorder()
			newTestRouter(t, svcMock).ServeHTTP(rec, req)

			assert.Equal(t, tc.expected, rec.Result().StatusCode, rec.Body.String())
		})
	}
}
