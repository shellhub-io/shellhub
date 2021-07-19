package routes

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"

	"github.com/shellhub-io/shellhub/api/apicontext"
	"github.com/shellhub-io/shellhub/api/services/mocks"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestAuthUser(t *testing.T) {
	e := echo.New()
	mock := new(mocks.Service)
	ctx := context.TODO()
	Err := errors.New("error")
	h := NewHandler(mock)

	var data models.UserAuthRequest

	t.Run("Authentication fails when the submited data is incorrect", func(t *testing.T) {
		authFail := models.UserAuthRequest{
			Username: "user1",
			Password: "wrong_password",
		}

		bytesData, _ := json.Marshal(authFail)

		rec := httptest.NewRecorder()

		req, err := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(bytesData))
		mock.On("AuthUser", ctx, authFail).Return(nil, Err)
		assert.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		echoContext := e.NewContext(req, rec)

		apictx := apicontext.NewContext(mock, echoContext)

		err = h.AuthUser(*apictx)

		assert.Equal(t, echo.ErrUnauthorized, err)
	})

	t.Run("Authentication succeeds", func(t *testing.T) {
		invalidUser := `{"username": "user1", "password": "password"}`

		authResponse := &models.UserAuthResponse{
			Token:  "token",
			User:   "user",
			Name:   "user1",
			ID:     "userID",
			Tenant: "tenant",
			Email:  "user@email.com",
		}

		_ = json.Unmarshal([]byte(invalidUser), &data)

		rr := httptest.NewRecorder()

		req, err := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer([]byte(invalidUser)))
		assert.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		echoContext := e.NewContext(req, rr)

		mock.On("AuthUser", ctx, data).Return(authResponse, nil)

		apictx := apicontext.NewContext(mock, echoContext)

		_ = h.AuthUser(*apictx)
		assert.Equal(t, http.StatusOK, rr.Code)
	})
}
