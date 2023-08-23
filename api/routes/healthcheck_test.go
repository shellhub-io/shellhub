package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/shellhub-io/shellhub/api/pkg/gateway"
	"github.com/shellhub-io/shellhub/api/services/mocks"
	"github.com/stretchr/testify/assert"
)

func TestEvaluateHealth(t *testing.T) {
	e := echo.New()
	mock := new(mocks.Service)
	h := NewHandler(mock)

	cases := []struct {
		title         string
		requiredMocks func()
		expectedErr   error
	}{
		{
			title:       "success when try to make a evaluate health",
			expectedErr: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.title, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, HealthCheckURL, nil)
			rec := httptest.NewRecorder()
			echoContext := e.NewContext(req, rec)

			apictx := gateway.NewContext(mock, echoContext)
			err := h.EvaluateHealth(*apictx)

			assert.Equal(t, tc.expectedErr, err)
			assert.Equal(t, http.StatusOK, rec.Code)
		})
	}

	mock.AssertExpectations(t)
}
