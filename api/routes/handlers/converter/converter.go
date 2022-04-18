package converter

import (
	"net/http"

	"github.com/shellhub-io/shellhub/api/services"
)

// FromErrServiceToHTTPStatus converts an service error code to http status.
func FromErrServiceToHTTPStatus(code int) int {
	switch code {
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
	default:
		return http.StatusInternalServerError
	}
}
