package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/shellhub-io/shellhub/api/pkg/guard"
	svc "github.com/shellhub-io/shellhub/api/services"
	"github.com/shellhub-io/shellhub/api/services/mocks"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/stretchr/testify/assert"
	gomock "github.com/stretchr/testify/mock"
)

func TestUpdateUserData(t *testing.T) {
	mock := new(mocks.Service)

	cases := []struct {
		title             string
		uid               string
		updatePayloadMock requests.UserDataUpdate
		requiredMocks     func(updatePayloadMock requests.UserDataUpdate)
		expectedStatus    int
	}{
		{
			title: "returns Ok when updating an existing user",
			uid:   "123",
			updatePayloadMock: requests.UserDataUpdate{
				UserParam: requests.UserParam{
					ID: "123",
				},
				Name:     "new name",
				Username: "usernameteste",
				Email:    "newemail@example.com",
			},
			requiredMocks: func(updatePayloadMock requests.UserDataUpdate) {
				mock.On("UpdateDataUser", gomock.Anything, "123", updatePayloadMock).Return(nil, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			title: "Update UserData not found",
			uid:   "1234",
			updatePayloadMock: requests.UserDataUpdate{
				UserParam: requests.UserParam{
					ID: "1234",
				},
				Name:     "new name",
				Username: "usernameteste",
				Email:    "newemail@example.com",
			},
			requiredMocks: func(updatePayloadMock requests.UserDataUpdate) {
				mock.On("UpdateDataUser", gomock.Anything, "1234", updatePayloadMock).Return(nil, svc.ErrUserNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tc := range cases {
		t.Run(tc.title, func(t *testing.T) {
			tc.requiredMocks(tc.updatePayloadMock)

			jsonData, err := json.Marshal(tc.updatePayloadMock)
			if err != nil {
				assert.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/users/%s/data", tc.uid), strings.NewReader(string(jsonData)))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Role", guard.RoleOwner)
			rec := httptest.NewRecorder()

			e := NewRouter(mock)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatus, rec.Result().StatusCode)
		})
	}

	mock.AssertExpectations(t)
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
			title: "returns Ok when updating a password an existing user",
			uid:   "123",
			updatePayloadMock: requests.UserPasswordUpdate{
				UserParam: requests.UserParam{
					ID: "123",
				},
				CurrentPassword: "old_password",
				NewPassword:     "new_password",
			},
			requiredMocks: func(updatePayloadMock requests.UserPasswordUpdate) {
				mock.On("UpdatePasswordUser", gomock.Anything, "123", updatePayloadMock.CurrentPassword, updatePayloadMock.NewPassword).Return(nil)
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
			req.Header.Set("X-Role", guard.RoleOwner)
			rec := httptest.NewRecorder()

			e := NewRouter(mock)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatus, rec.Result().StatusCode)
		})
	}

	mock.AssertExpectations(t)
}
