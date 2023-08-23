package routes

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	svc "github.com/shellhub-io/shellhub/api/services"

	"github.com/shellhub-io/shellhub/api/store"

	"github.com/shellhub-io/shellhub/api/pkg/guard"
	"github.com/shellhub-io/shellhub/api/services/mocks"
	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	gomock "github.com/stretchr/testify/mock"
)

func TestGetSessionList(t *testing.T) {
	mock := new(mocks.Service)

	type Expected struct {
		expectedSession []models.Session
		expectedStatus  int
	}
	cases := []struct {
		title         string
		payload       paginator.Query
		requiredMocks func(payload *paginator.Query)
		expected      Expected
	}{
		{
			title: "fails when try to searching a session list of a existing session",
			payload: paginator.Query{
				Page:    1,
				PerPage: 10,
			},
			requiredMocks: func(payload *paginator.Query) {
				mock.On("ListSessions", gomock.Anything, *payload).Return(nil, 0, svc.ErrNotFound).Once()
			},
			expected: Expected{
				expectedSession: nil,
				expectedStatus:  http.StatusNotFound,
			},
		},
		{
			title: "success when try to searching a session list of a existing session",
			payload: paginator.Query{
				Page:    1,
				PerPage: 10,
			},
			requiredMocks: func(payload *paginator.Query) {
				ss := []models.Session{}
				mock.On("ListSessions", gomock.Anything, *payload).Return(ss, 1, nil).Once()
			},
			expected: Expected{
				expectedSession: []models.Session{},
				expectedStatus:  http.StatusOK,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.title, func(t *testing.T) {
			tc.requiredMocks(&tc.payload)

			jsonData, err := json.Marshal(tc.payload)
			if err != nil {
				assert.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodGet, "/api/sessions", strings.NewReader(string(jsonData)))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Role", guard.RoleOwner)
			rec := httptest.NewRecorder()

			e := NewRouter(mock)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expected.expectedStatus, rec.Result().StatusCode)

			var session []models.Session
			if err := json.NewDecoder(rec.Result().Body).Decode(&session); err != nil {
				assert.ErrorIs(t, io.EOF, err)
			}
			assert.Equal(t, tc.expected.expectedSession, session)
		})
	}

	mock.AssertExpectations(t)
}

func TestGetSession(t *testing.T) {
	mock := new(mocks.Service)

	type Expected struct {
		expectedSession *models.Session
		expectedStatus  int
	}
	cases := []struct {
		title         string
		uid           string
		requiredMocks func(session *models.Session)
		expected      Expected
	}{
		{
			title:         "fails when try to get session don't existing",
			uid:           "",
			requiredMocks: func(*models.Session) {},
			expected: Expected{
				expectedSession: nil,
				expectedStatus:  http.StatusNotFound,
			},
		},
		{
			title: "fails when try to get session don't existing",
			uid:   "1234",
			requiredMocks: func(*models.Session) {
				mock.On("GetSession", gomock.Anything, models.UID("1234")).Return(nil, svc.NewErrSessionNotFound(models.UID("1234"), store.ErrNoDocuments))
			},
			expected: Expected{
				expectedSession: nil,
				expectedStatus:  http.StatusNotFound,
			},
		},
		{
			title: "success when try to get a session exists",
			uid:   "123",
			requiredMocks: func(session *models.Session) {
				mock.On("GetSession", gomock.Anything, models.UID("123")).Return(session, nil)
			},
			expected: Expected{
				expectedSession: &models.Session{UID: "123"},
				expectedStatus:  http.StatusOK,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.title, func(t *testing.T) {
			tc.requiredMocks(tc.expected.expectedSession)

			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/sessions/%s", tc.uid), nil)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Role", guard.RoleOwner)
			rec := httptest.NewRecorder()

			e := NewRouter(mock)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expected.expectedStatus, rec.Result().StatusCode)

			var session *models.Session
			if err := json.NewDecoder(rec.Result().Body).Decode(&session); err != nil {
				assert.ErrorIs(t, io.EOF, err)
			}

			assert.Equal(t, tc.expected.expectedSession, session)
		})
	}

	mock.AssertExpectations(t)
}

func TestCreateSession(t *testing.T) {
	mock := new(mocks.Service)

	cases := []struct {
		title          string
		request        requests.SessionCreate
		requiredMocks  func()
		expectedStatus int
	}{
		{
			title: "fails when bind fails to validate uid",
			request: requests.SessionCreate{
				DeviceUID: "xyz789",
				Username:  "johndoe",
				IPAddress: "192.168.0.1",
				Type:      "session",
				Term:      "2023Q2",
			},
			requiredMocks:  func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			title: "fails when try to creating a non-existing session",
			request: requests.SessionCreate{
				UID:       "1234",
				DeviceUID: "xyz789",
				Username:  "johndoe",
				IPAddress: "192.168.0.1",
				Type:      "session",
				Term:      "2023Q2",
			},
			requiredMocks: func() {
				mock.On("CreateSession", gomock.Anything, requests.SessionCreate{
					UID:       "1234",
					DeviceUID: "xyz789",
					Username:  "johndoe",
					IPAddress: "192.168.0.1",
					Type:      "session",
					Term:      "2023Q2",
				},
				).Return(nil, svc.ErrSessionNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			title: "success when try to creating an existing session",
			request: requests.SessionCreate{
				UID:       "123",
				DeviceUID: "xyz789",
				Username:  "johndoe",
				IPAddress: "192.168.0.1",
				Type:      "session",
				Term:      "2023Q2",
			},
			requiredMocks: func() {
				mock.On("CreateSession", gomock.Anything, requests.SessionCreate{
					UID:       "123",
					DeviceUID: "xyz789",
					Username:  "johndoe",
					IPAddress: "192.168.0.1",
					Type:      "session",
					Term:      "2023Q2",
				},
				).Return(&models.Session{}, nil)
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range cases {
		t.Run(tc.title, func(t *testing.T) {
			tc.requiredMocks()

			jsonData, err := json.Marshal(tc.request)
			if err != nil {
				assert.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPost, "/internal/sessions", strings.NewReader(string(jsonData)))
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

func TestFinishSession(t *testing.T) {
	mock := new(mocks.Service)

	cases := []struct {
		title          string
		uid            string
		requiredMocks  func()
		expectedStatus int
	}{
		{
			title:          "fails when bind fails to validate uid",
			uid:            "",
			requiredMocks:  func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			title: "fails when try to finishing a non-existing session",
			uid:   "1234",
			requiredMocks: func() {
				mock.On("DeactivateSession", gomock.Anything, models.UID("1234")).Return(svc.ErrSessionNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			title: "success when try to finishing an existing session",
			uid:   "123",
			requiredMocks: func() {
				mock.On("DeactivateSession", gomock.Anything, models.UID("123")).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range cases {
		t.Run(tc.title, func(t *testing.T) {
			tc.requiredMocks()

			req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/internal/sessions/%s/finish", tc.uid), nil)
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
