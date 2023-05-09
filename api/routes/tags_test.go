package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/shellhub-io/shellhub/api/pkg/guard"
	"github.com/shellhub-io/shellhub/api/services/mocks"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/stretchr/testify/assert"
	gomock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetTags(t *testing.T) {
	mock := new(mocks.Service)

	cases := []struct {
		title          string
		requiredMocks  func()
		expectedStatus int
	}{
		{
			title:          "returns Ok if a tags exists",
			expectedStatus: http.StatusOK,
			requiredMocks: func() {
				mock.On("GetTags", gomock.Anything, "").Return([]string{"tag1", "tag2"}, 2, nil)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.title, func(t *testing.T) {
			tc.requiredMocks()

			req := httptest.NewRequest(http.MethodGet, "/api/tags", nil)
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

func TestRenameTag(t *testing.T) {
	mock := new(mocks.Service)

	cases := []struct {
		title          string
		requiredMocks  func()
		expectedStatus int
		expectedTags   requests.TagRename
	}{
		{
			title: "returns Ok when renaming an existing tag",
			expectedTags: requests.TagRename{
				TagParam: requests.TagParam{Tag: "oldTag"},
				NewTag:   "newTag",
			},
			expectedStatus: http.StatusOK,
			requiredMocks: func() {
				mock.On("RenameTag", gomock.Anything, "", "oldTag", "newTag").Return(nil)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.title, func(t *testing.T) {
			tc.requiredMocks()
			jsonData, err := json.Marshal(tc.expectedTags)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/tags/%s", tc.expectedTags.Tag), strings.NewReader(string(jsonData)))
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

func TestDeleteTag(t *testing.T) {
	mock := new(mocks.Service)

	cases := []struct {
		title          string
		requiredMocks  func()
		expectedStatus int
		expectedTags   requests.TagDelete
		tenant         string
	}{
		{
			title: "returns Ok when deleting an existing tag",
			expectedTags: requests.TagDelete{
				TagParam: requests.TagParam{Tag: "tagtest"},
			},
			tenant:         "tenant",
			expectedStatus: http.StatusOK,
			requiredMocks: func() {
				mock.On("DeleteTag", gomock.Anything, "tenant", "tagtest").Return(nil)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.title, func(t *testing.T) {
			tc.requiredMocks()
			jsonData, err := json.Marshal(tc.expectedTags)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/tags/%s", tc.expectedTags), strings.NewReader(string(jsonData)))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Role", guard.RoleOwner)
			req.Header.Set("X-Tenant-ID", tc.tenant)
			rec := httptest.NewRecorder()

			e := NewRouter(mock)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatus, rec.Result().StatusCode)
		})
	}

	mock.AssertExpectations(t)
}
