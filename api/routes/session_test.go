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

	cases := []struct {
		title           string
		payload         paginator.Query
		requiredMocks   func(payload *paginator.Query)
		expectedSession []models.Session
		expectedStatus  int
	}{
		{
			title: "returns Ok when searching a session list of a existing session",
			payload: paginator.Query{
				Page:    1,
				PerPage: 10,
			},
			requiredMocks: func(payload *paginator.Query) {
				ss := []models.Session{}
				mock.On("ListSessions", gomock.Anything, *payload).Return(ss, 1, nil)
			},
			expectedSession: []models.Session{},
			expectedStatus:  http.StatusOK,
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

			assert.Equal(t, tc.expectedStatus, rec.Result().StatusCode)

			var session []models.Session
			if err := json.NewDecoder(rec.Result().Body).Decode(&session); err != nil {
				assert.ErrorIs(t, io.EOF, err)
			}
			assert.Equal(t, tc.expectedSession, session)
		})
	}

	mock.AssertExpectations(t)
}

func TestGetSession(t *testing.T) {
	mock := new(mocks.Service)

	cases := []struct {
		title           string
		uid             string
		requiredMocks   func(session *models.Session)
		expectedSession *models.Session
		expectedStatus  int
	}{
		{
			title:           "returns Ok if a session exists",
			uid:             "123",
			expectedSession: &models.Session{UID: "123"},
			requiredMocks: func(session *models.Session) {
				mock.On("GetSession", gomock.Anything, models.UID("123")).Return(session, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			title:           "returns Not Found if a session don't existing",
			uid:             "1234",
			expectedSession: nil,
			requiredMocks: func(*models.Session) {
				mock.On("GetSession", gomock.Anything, models.UID("1234")).Return(nil, svc.NewErrSessionNotFound(models.UID("1234"), store.ErrNoDocuments))
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tc := range cases {
		t.Run(tc.title, func(t *testing.T) {
			tc.requiredMocks(tc.expectedSession)

			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/sessions/%s", tc.uid), nil)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Role", guard.RoleOwner)
			rec := httptest.NewRecorder()

			e := NewRouter(mock)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatus, rec.Result().StatusCode)

			var session *models.Session
			if err := json.NewDecoder(rec.Result().Body).Decode(&session); err != nil {
				assert.ErrorIs(t, io.EOF, err)
			}

			assert.Equal(t, tc.expectedSession, session)
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
			title: "returns Ok when creating an existing session",
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
		{
			title: "returns Not Found when creating a non-existing session",
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
			title: "returns Ok when finishing an existing session",
			uid:   "123",
			requiredMocks: func() {
				mock.On("DeactivateSession", gomock.Anything, models.UID("123")).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			title: "returns Not Found when finishing a non-existing session",
			uid:   "1234",
			requiredMocks: func() {
				mock.On("DeactivateSession", gomock.Anything, models.UID("1234")).Return(svc.ErrSessionNotFound)
			},
			expectedStatus: http.StatusNotFound,
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
