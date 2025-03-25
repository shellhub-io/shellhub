package routes

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/shellhub-io/shellhub/api/pkg/gateway"
	svc "github.com/shellhub-io/shellhub/api/services"
	"github.com/shellhub-io/shellhub/api/services/mocks"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
	websocketmocks "github.com/shellhub-io/shellhub/pkg/websocket/mocks"
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
		description   string
		paginator     query.Paginator
		requiredMocks func(paginator query.Paginator)
		expected      Expected
	}{
		{
			description: "fails when try to searching a session list of a existing session",
			paginator: query.Paginator{
				Page:    1,
				PerPage: 10,
			},
			requiredMocks: func(paginator query.Paginator) {
				mock.On("ListSessions", gomock.Anything, paginator).Return(nil, 0, svc.ErrNotFound).Once()
			},
			expected: Expected{
				expectedSession: nil,
				expectedStatus:  http.StatusNotFound,
			},
		},
		{
			description: "success when try to searching a session list of a existing session",
			paginator: query.Paginator{
				Page:    1,
				PerPage: 10,
			},
			requiredMocks: func(paginator query.Paginator) {
				ss := []models.Session{}
				mock.On("ListSessions", gomock.Anything, paginator).Return(ss, 1, nil).Once()
			},
			expected: Expected{
				expectedSession: []models.Session{},
				expectedStatus:  http.StatusOK,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks(tc.paginator)

			jsonData, err := json.Marshal(tc.paginator)
			if err != nil {
				assert.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodGet, "/api/sessions", strings.NewReader(string(jsonData)))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Role", authorizer.RoleOwner.String())
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
			req.Header.Set("X-Role", authorizer.RoleOwner.String())
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
			req.Header.Set("X-Role", authorizer.RoleOwner.String())
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
			req.Header.Set("X-Role", authorizer.RoleOwner.String())
			rec := httptest.NewRecorder()

			e := NewRouter(mock)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatus, rec.Result().StatusCode)
		})
	}

	mock.AssertExpectations(t)
}

func TestEventSession(t *testing.T) {
	mock := new(mocks.Service)
	webSocketUpgraderMock := new(websocketmocks.Upgrader)

	cases := []struct {
		description   string
		uid           string
		seat          int
		requiredMocks func(uid string)
		expected      int
	}{
		{
			description: "fails when upgrade cannot be done",
			uid:         "123",
			seat:        0,
			requiredMocks: func(_ string) {
				webSocketUpgraderMock.On("Upgrade", gomock.Anything, gomock.Anything).Return(nil, errors.New("")).Once()
			},
			expected: http.StatusBadRequest,
		},
		{
			description: "fails when cannot read from websocket due error",
			uid:         "123",
			seat:        0,
			requiredMocks: func(_ string) {
				conn := new(websocketmocks.Conn)
				conn.On("Close").Return(nil).Once()
				conn.On("ReadJSON", gomock.Anything).Return(io.EOF).Once()

				webSocketUpgraderMock.On("Upgrade", gomock.Anything, gomock.Anything).Return(conn, nil).Once()
			},
			expected: http.StatusInternalServerError,
		},
		{
			description: "fails when cannot read from websocket due generic error",
			uid:         "123",
			seat:        0,
			requiredMocks: func(_ string) {
				conn := new(websocketmocks.Conn)
				conn.On("Close").Return(nil).Once()
				conn.On("ReadJSON", gomock.Anything).Return(errors.New("")).Once()

				webSocketUpgraderMock.On("Upgrade", gomock.Anything, gomock.Anything).Return(conn, nil).Once()
			},
			expected: http.StatusInternalServerError,
		},
		{
			description: "fails when record frame is invalid",
			uid:         "123",
			seat:        0,
			requiredMocks: func(_ string) {
				conn := new(websocketmocks.Conn)
				conn.On("Close").Return(nil).Once()
				conn.On("ReadJSON", gomock.Anything).Return(nil).Once().Run(func(args gomock.Arguments) {
					req := args.Get(0).(*requests.SessionEvent) //nolint:forcetypeassert

					json.
						NewDecoder(strings.NewReader(`{}`)).
						Decode(req) //nolint:errcheck
				})

				webSocketUpgraderMock.On("Upgrade", gomock.Anything, gomock.Anything).Return(conn, nil).Once()
			},
			expected: http.StatusBadRequest,
		},
		{
			description: "fails to write the frame on the database",
			uid:         "123",
			seat:        0,
			requiredMocks: func(uid string) {
				conn := new(websocketmocks.Conn)
				conn.On("Close").Return(nil).Once()
				conn.On("NextReader").Return().Once()
				conn.On("ReadJSON", gomock.Anything).Return(nil).Once().Run(func(args gomock.Arguments) {
					req := args.Get(0).(*requests.SessionEvent) //nolint:forcetypeassert

					json.
						NewDecoder(strings.NewReader(`{"type":"pty-output","timestamp":"2025-02-03T14:11:32.405Z","data": { "output":"test" },"seat": 0}`)).
						Decode(req) //nolint:errcheck
				})

				webSocketUpgraderMock.On("Upgrade", gomock.Anything, gomock.Anything).Return(conn, nil).Once()

				mock.On("EventSession", gomock.Anything, models.UID(uid), gomock.Anything).
					Return(errors.New("not able record")).Once()
			},
			expected: http.StatusInternalServerError,
		},
		{
			description: "success to write one frame on database",
			uid:         "123",
			seat:        0,
			requiredMocks: func(uid string) {
				conn := new(websocketmocks.Conn)
				conn.On("Close").Return(nil).Once()
				conn.On("NextReader").Return().Once()
				conn.On("ReadJSON", gomock.Anything).Return(nil).Once().Run(func(args gomock.Arguments) {
					req := args.Get(0).(*requests.SessionEvent) //nolint:forcetypeassert

					json.
						NewDecoder(strings.NewReader(`{"type":"pty-output","timestamp":"2025-02-03T14:11:32.405Z","data": { "output":"test" },"seat": 0}`)).
						Decode(req) //nolint:errcheck
				})

				webSocketUpgraderMock.On("Upgrade", gomock.Anything, gomock.Anything).Return(conn, nil).Once()

				mock.On("EventSession", gomock.Anything, models.UID(uid),
					gomock.Anything).Return(nil).Once()

				conn.On("ReadJSON", gomock.Anything).Return(&websocket.CloseError{
					Code: 1000,
					Text: "test",
				}).Once()
			},
			expected: http.StatusOK,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks(tc.uid)

			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("ws:///internal/sessions/%s/events", tc.uid), nil)
			req.Header.Set("Content-Type", echo.MIMEApplicationJSON)
			req.Header.Set("X-Role", authorizer.RoleOwner.String())
			req.Header.Set("Upgrade", "websocket")
			req.Header.Set("Connection", "Upgrade")
			req.Header.Set("Sec-WebSocket-Version", "13")
			req.Header.Set("Sec-WebSocket-Key", "test")

			e := NewRouter(mock, func(_ *echo.Echo, handler *Handler) error {
				handler.WebSocketUpgrader = webSocketUpgraderMock

				return nil
			})

			e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
				return func(c echo.Context) error {
					ctx := gateway.NewContext(mock, c)

					return next(ctx)
				}
			})

			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expected, rec.Result().StatusCode)
			mock.AssertExpectations(t)
		})
	}
}
