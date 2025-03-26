package converter

import (
	"net/http"

	routes "github.com/shellhub-io/shellhub/server/api/routes/errors"
	services "github.com/shellhub-io/shellhub/server/api/services"
)

// FromErrServiceToHTTPStatus converts a service error code to http status.
func FromErrServiceToHTTPStatus(code int) int {
	switch code {
	case services.ErrCodeCreated:
		return http.StatusCreated
	case services.ErrCodeNotFound:
		return http.StatusNotFound
	case services.ErrCodeInvalid:
		return http.StatusBadRequest
	case services.ErrCodeLimit:
		return http.StatusForbidden
	case services.ErrCodePayment:
		return http.StatusPaymentRequired
	case services.ErrCodeDuplicated:
		return http.StatusConflict
	case services.ErrCodeUnauthorized:
		return http.StatusUnauthorized
	case services.ErrCodeForbidden:
		return http.StatusForbidden
	case services.ErrCodeNoContentChange:
		return http.StatusNoContent
	default:
		return http.StatusInternalServerError
	}
}

// FromErrRouteToHTTPStatus converts a route error code to http status.
func FromErrRouteToHTTPStatus(code int) int {
	switch code {
	case routes.ErrCodeUnprocessableEntity:
		return http.StatusUnprocessableEntity
	case routes.ErrCodeInvalidEntity:
		return http.StatusBadRequest
	case routes.ErrCodeUnauthorized:
		return http.StatusUnauthorized
	default:
		return http.StatusInternalServerError
	}
}
