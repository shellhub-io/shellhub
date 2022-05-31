package routes

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/shellhub-io/shellhub/api/pkg/gateway"
	"github.com/shellhub-io/shellhub/api/pkg/guard"
	svc "github.com/shellhub-io/shellhub/api/services"
	"github.com/shellhub-io/shellhub/api/services/mocks"
	"github.com/shellhub-io/shellhub/pkg/errors"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestUpdatePendingStatus(t *testing.T) {
	e := echo.New()
	mock := new(mocks.Service)
	ctx := context.TODO()
	h := NewHandler(mock)

	fromErrServiceToHTTPStatus := func(code int) int {
		switch code {
		case svc.ErrCodeNotFound:
			return http.StatusNotFound
		case svc.ErrCodeInvalid:
			return http.StatusBadRequest
		case svc.ErrCodeLimit:
			return http.StatusForbidden
		case svc.ErrCodePayment:
			return http.StatusPaymentRequired
		case svc.ErrCodeDuplicated:
			return http.StatusConflict
		case svc.ErrCodeUnauthorized:
			return http.StatusUnauthorized
		case svc.ErrCodeForbidden:
			return http.StatusForbidden
		default:
			return http.StatusInternalServerError
		}
	}

	t.Run("Return payment required when the max count is reached", func(t *testing.T) {
		rec := httptest.NewRecorder()

		req, _ := http.NewRequest(http.MethodPatch, "/devices/:uid/:status", bytes.NewBuffer([]byte{}))
		req.Header.Set("Content-Role", "application/json")
		req.Header.Set("X-Role", guard.RoleOwner)
		echoContext := e.NewContext(req, rec)
		echoContext.SetParamNames("uid", "status")
		echoContext.SetParamValues("123", "pending")
		mock.On("UpdatePendingStatus", ctx, models.UID("123"), "pending", "").Return(svc.ErrDeviceLimit)

		apictx := gateway.NewContext(mock, echoContext)

		output := h.UpdatePendingStatus(*apictx)
		err, ok := output.(errors.Error)
		assert.True(t, ok)

		assert.Equal(t, http.StatusPaymentRequired, fromErrServiceToHTTPStatus(err.Code))
	})
}
