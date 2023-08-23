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
			title:          "success when try to get an existing tag",
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

	type Expected struct {
		expectedTags   requests.TagRename
		expectedStatus int
	}
	cases := []struct {
		title         string
		requiredMocks func()
		expected      Expected
	}{
		{
			title: "fails when bind fails to validate uid",
			expected: Expected{
				expectedTags: requests.TagRename{
					TagParam: requests.TagParam{Tag: "oldTag"},
				},
				expectedStatus: http.StatusBadRequest,
			},
			requiredMocks: func() {},
		},
		{
			title: "fails when validate because the tag does not have a min of 3 characters",
			expected: Expected{
				expectedTags: requests.TagRename{
					TagParam: requests.TagParam{Tag: "oldTag"},
					NewTag:   "tg",
				},
				expectedStatus: http.StatusBadRequest,
			},
			requiredMocks: func() {},
		},
		{
			title: "fails when validate because the tag does not have a max of 255 characters",
			expected: Expected{
				expectedTags: requests.TagRename{
					TagParam: requests.TagParam{Tag: "oldTag"},
					NewTag:   "BCD3821E12F7A6D89295D86E277F2C365D7A4C3FCCD75D8A2F46C0A556A8EBAAF0845C85D50241FC2F9806D8668FF75D262FDA0A055784AD36D8CA7D2BB600C9BCD3821E12F7A6D89295D86E277F2C365D7A4C3FCCD75D8A2F46C0A556A8EBAAF0845C85D50241FC2F9806D8668FF75D262FDA0A055784AD36D8CA7D2BB600C9",
				},
				expectedStatus: http.StatusBadRequest,
			},
			requiredMocks: func() {},
		},
		{
			title: "fails when validate because have a '/' with in your characters",
			expected: Expected{
				expectedTags: requests.TagRename{
					TagParam: requests.TagParam{Tag: "oldTag"},
					NewTag:   "/",
				},
				expectedStatus: http.StatusBadRequest,
			},
			requiredMocks: func() {},
		},
		{
			title: "fails when validate because have a '&' with in your characters",
			expected: Expected{
				expectedTags: requests.TagRename{
					TagParam: requests.TagParam{Tag: "oldTag"},
					NewTag:   "&",
				},
				expectedStatus: http.StatusBadRequest,
			},
			requiredMocks: func() {},
		},
		{
			title: "fails when validate because have a '@' with in your characters",
			expected: Expected{
				expectedTags: requests.TagRename{
					TagParam: requests.TagParam{Tag: "oldTag"},
					NewTag:   "@",
				},
				expectedStatus: http.StatusBadRequest,
			},
			requiredMocks: func() {},
		},
		{
			title: "success when try to renaming an existing tag",
			expected: Expected{
				expectedTags: requests.TagRename{
					TagParam: requests.TagParam{Tag: "oldTag"},
					NewTag:   "newTag",
				},
				expectedStatus: http.StatusOK,
			},
			requiredMocks: func() {
				mock.On("RenameTag", gomock.Anything, "", "oldTag", "newTag").Return(nil)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.title, func(t *testing.T) {
			tc.requiredMocks()
			jsonData, err := json.Marshal(tc.expected.expectedTags)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/tags/%s", tc.expected.expectedTags.Tag), strings.NewReader(string(jsonData)))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Role", guard.RoleOwner)
			rec := httptest.NewRecorder()

			e := NewRouter(mock)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expected.expectedStatus, rec.Result().StatusCode)
		})
	}

	mock.AssertExpectations(t)
}

func TestDeleteTag(t *testing.T) {
	mock := new(mocks.Service)

	type Expected struct {
		expectedTags   requests.TagDelete
		expectedStatus int
	}
	cases := []struct {
		title         string
		requiredMocks func()
		tenant        string
		expected      Expected
	}{
		{
			title: "fails when bind fails to validate uid",
			expected: Expected{
				expectedTags: requests.TagDelete{
					TagParam: requests.TagParam{Tag: ""},
				},

				expectedStatus: http.StatusBadRequest,
			},
			requiredMocks: func() {},
		},
		{
			title: "fails when validate because the tag does not have a min of 3 characters",
			expected: Expected{
				expectedTags: requests.TagDelete{
					TagParam: requests.TagParam{Tag: "tg"},
				},

				expectedStatus: http.StatusBadRequest,
			},
			requiredMocks: func() {},
		},
		{
			title: "fails when validate because the tag does not have a max of 255 characters",
			expected: Expected{
				expectedTags: requests.TagDelete{
					TagParam: requests.TagParam{Tag: "BCD3821E12F7A6D89295D86E277F2C365D7A4C3FCCD75D8A2F46C0A556A8EBAAF0845C85D50241FC2F9806D8668FF75D262FDA0A055784AD36D8CA7D2BB600C9BCD3821E12F7A6D89295D86E277F2C365D7A4C3FCCD75D8A2F46C0A556A8EBAAF0845C85D50241FC2F9806D8668FF75D262FDA0A055784AD36D8CA7D2BB600C9"},
				},

				expectedStatus: http.StatusBadRequest,
			},
			requiredMocks: func() {},
		},
		{
			title: "fails when validate because have a '/' with in your characters",
			expected: Expected{
				expectedTags: requests.TagDelete{
					TagParam: requests.TagParam{Tag: "/"},
				},

				expectedStatus: http.StatusBadRequest,
			},
			requiredMocks: func() {},
		},
		{
			title: "fails when validate because have a '&' with in your characters",
			expected: Expected{
				expectedTags: requests.TagDelete{
					TagParam: requests.TagParam{Tag: "&"},
				},

				expectedStatus: http.StatusBadRequest,
			},
			requiredMocks: func() {},
		},
		{
			title: "fails when validate because have a '@' with in your characters",
			expected: Expected{
				expectedTags: requests.TagDelete{
					TagParam: requests.TagParam{Tag: "@"},
				},

				expectedStatus: http.StatusBadRequest,
			},
			requiredMocks: func() {},
		},
		{
			title: "success when try to deleting an existing tag",
			expected: Expected{
				expectedTags: requests.TagDelete{
					TagParam: requests.TagParam{Tag: "tagtest"},
				},
				expectedStatus: http.StatusOK,
			},
			tenant: "tenant",
			requiredMocks: func() {
				mock.On("DeleteTag", gomock.Anything, "tenant", "tagtest").Return(nil)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.title, func(t *testing.T) {
			tc.requiredMocks()
			jsonData, err := json.Marshal(tc.expected.expectedTags)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/tags/%s", tc.expected.expectedTags), strings.NewReader(string(jsonData)))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Role", guard.RoleOwner)
			req.Header.Set("X-Tenant-ID", tc.tenant)
			rec := httptest.NewRecorder()

			e := NewRouter(mock)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expected.expectedStatus, rec.Result().StatusCode)
		})
	}

	mock.AssertExpectations(t)
}
