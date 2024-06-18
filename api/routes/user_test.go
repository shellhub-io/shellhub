package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	svc "github.com/shellhub-io/shellhub/api/services"
	"github.com/shellhub-io/shellhub/api/services/mocks"
	"github.com/shellhub-io/shellhub/pkg/api/auth"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/stretchr/testify/assert"
	gomock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUpdateUserData(t *testing.T) {
	type Expected struct {
		status int
	}

	svcMock := new(mocks.Service)

	cases := []struct {
		description   string
		headers       map[string]string
		body          requests.UserDataUpdate
		requiredMocks func()
		expected      Expected
	}{
		{
			description: "fails when bind fails to validate e-mail",
			headers: map[string]string{
				"X-ID":   "000000000000000000000000",
				"X-Role": "owner",
			},
			body: requests.UserDataUpdate{
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
			body: requests.UserDataUpdate{
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
			body: requests.UserDataUpdate{
				Name:          "John Doe",
				Username:      "john_doe",
				Email:         "john.doe@test.com",
				RecoveryEmail: "john.doe@test.com",
			},
			requiredMocks: func() {
				svcMock.
					On(
						"UpdateDataUser",
						gomock.Anything,
						"000000000000000000000000",
						&requests.UserDataUpdate{
							Name:          "John Doe",
							Username:      "john_doe",
							Email:         "john.doe@test.com",
							RecoveryEmail: "john.doe@test.com",
						},
					).
					Return(nil, svc.ErrUserNotFound).
					Once()
			},
			expected: Expected{http.StatusNotFound},
		},
		{
			description: "success when try to updating an existing user",
			body: requests.UserDataUpdate{
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
						"UpdateDataUser",
						gomock.Anything,
						"000000000000000000000000",
						&requests.UserDataUpdate{
							Name:          "John Doe",
							Username:      "john_doe",
							Email:         "john.doe@test.com",
							RecoveryEmail: "john.doe@test.com",
						},
					).
					Return(nil, nil).
					Once()
			},
			expected: Expected{http.StatusOK},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/users/%s/data", tc.headers["X-ID"]), strings.NewReader(string(data)))
			req.Header.Set("Content-Type", "application/json")
			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}

			rec := httptest.NewRecorder()

			e := NewRouter(svcMock)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expected, Expected{rec.Result().StatusCode})
		})
	}

	svcMock.AssertExpectations(t)
}

func TestUpdateUserPassword(t *testing.T) {
	mock := new(mocks.Service)

	cases := []struct {
		title             string
		uid               string
		updatePayloadMock requests.UserPasswordUpdate
		requiredMocks     func(updatePayloadMock requests.UserPasswordUpdate)
		expectedStatus    int
	}{
		{
			title: "fails when bind fails to validate uid",
			uid:   "123",
			updatePayloadMock: requests.UserPasswordUpdate{
				UserParam: requests.UserParam{
					ID: "123",
				},
			},
			requiredMocks:  func(updatePayloadMock requests.UserPasswordUpdate) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			title: "fails when validate because the tag does not have a min of 5 characters",
			uid:   "123",
			updatePayloadMock: requests.UserPasswordUpdate{
				UserParam: requests.UserParam{
					ID: "123",
				},
				CurrentPassword: "fail",
				NewPassword:     "new_password",
			},
			requiredMocks:  func(updatePayloadMock requests.UserPasswordUpdate) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			title: "fails when validate because the tag does not have a max of 32 characters",
			uid:   "123",
			updatePayloadMock: requests.UserPasswordUpdate{
				UserParam: requests.UserParam{
					ID: "123",
				},
				CurrentPassword: "1a3b8f0c2e5d7g9i4k6m8o2q5s7u9w1v7",
				NewPassword:     "new_password",
			},
			requiredMocks:  func(updatePayloadMock requests.UserPasswordUpdate) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			title: "fails when validate because the tag does not have a min of 5 characters",
			uid:   "123",
			updatePayloadMock: requests.UserPasswordUpdate{
				UserParam: requests.UserParam{
					ID: "123",
				},
				CurrentPassword: "new_password",
				NewPassword:     "fail",
			},
			requiredMocks:  func(updatePayloadMock requests.UserPasswordUpdate) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			title: "fails when validate because the tag does not have a max of 32 characters",
			uid:   "123",
			updatePayloadMock: requests.UserPasswordUpdate{
				UserParam: requests.UserParam{
					ID: "123",
				},
				CurrentPassword: "new_password",
				NewPassword:     "1a3b8f0c2e5d7g9i4k6m8o2q5s7u9w1v7",
			},
			requiredMocks:  func(updatePayloadMock requests.UserPasswordUpdate) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			title: "fails when validate because have a duplicate password",
			uid:   "123",
			updatePayloadMock: requests.UserPasswordUpdate{
				UserParam: requests.UserParam{
					ID: "123",
				},
				NewPassword:     "duplicate",
				CurrentPassword: "duplicate",
			},
			requiredMocks:  func(updatePayloadMock requests.UserPasswordUpdate) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			title: "fails when try to updating a password an existing user",
			uid:   "123",
			updatePayloadMock: requests.UserPasswordUpdate{
				UserParam: requests.UserParam{
					ID: "123",
				},
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
			uid:   "123",
			updatePayloadMock: requests.UserPasswordUpdate{
				UserParam: requests.UserParam{
					ID: "123",
				},
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
				assert.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/users/%s/password", tc.uid), strings.NewReader(string(jsonData)))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Role", auth.RoleOwner.String())
			rec := httptest.NewRecorder()

			e := NewRouter(mock)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatus, rec.Result().StatusCode)
		})
	}

	mock.AssertExpectations(t)
}
