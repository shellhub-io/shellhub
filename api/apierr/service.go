package apierr

import (
	"errors"
	"net/http"

	"github.com/shellhub-io/shellhub/api/apicontext"
)

var (
	ErrUnauthorized     = errors.New("unauthorized")
	ErrUserNotFound     = errors.New("user not found")
	ErrResourceNotFound = errors.New("resource not found")
	ErrDuplicateID      = errors.New("user already member of this namespace")
	ErrUserOwner        = errors.New("cannot remove this user")
)

func HandleError(c apicontext.Context, err error) error {
	switch err {
	case ErrUnauthorized:
		return c.NoContent(http.StatusForbidden)
	case ErrUserNotFound:
		return c.String(http.StatusNotFound, err.Error())
	case ErrResourceNotFound:
		return c.String(http.StatusNotFound, err.Error())
	case ErrDuplicateID:
		return c.String(http.StatusConflict, err.Error())
	case ErrUserOwner:
		return c.NoContent(http.StatusForbidden)
	default:
		return err
	}
}
