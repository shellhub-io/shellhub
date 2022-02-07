package services

import (
	"errors"
	"fmt"

	"github.com/shellhub-io/shellhub/api/store"
)

func newError(prefix string, err error) error {
	return fmt.Errorf("%s: %w", prefix, err)
}

var (
	ErrNamespaceMemberNotFound   = newError("namespace", errors.New("namespace: member not found"))
	ErrNamespaceDuplicatedMember = newError("namespace", errors.New("this member already exist in this namespace"))
	ErrNamespaceNameInvalid      = newError("namespace", errors.New("name is invalid"))
	ErrNamespaceTenantInvalid    = newError("namespace", errors.New("tenant is invalid"))
	ErrNamespaceRename           = newError("namespace", store.ErrNamespaceRename)
	ErrNamespaceNotFound         = newError("namespace", store.ErrNamespaceNotFound)
)

var (
	ErrNotFound              = errors.New("not found")
	ErrConflict              = errors.New("conflict")
	ErrBadRequest            = errors.New("bad request")
	ErrUnauthorized          = errors.New("unauthorized")
	ErrForbidden             = errors.New("forbidden")
	ErrUserNotFound          = errors.New("user not found")
	ErrDuplicateID           = errors.New("user already member of this namespace")
	ErrConflictName          = errors.New("this Name already exists")
	ErrInvalidFormat         = errors.New("invalid Name format")
	ErrMaxDeviceCountReached = errors.New("maximum number of accepted devices reached")
	ErrDuplicatedDeviceName  = errors.New("the Name already exists in the namespace")
	ErrDuplicateFingerprint  = errors.New("this fingerprint already exits")
	ErrReport                = errors.New("report error")
	ErrMaxTagReached         = errors.New("maximum number of tags reached")
	ErrDeviceNotFound        = errors.New("device not found")
	ErrDuplicateTagName      = errors.New("this tag Name already exists")
	ErrTagNameNotFound       = errors.New("tag Name not found")
	ErrNoTags                = errors.New("no tags has found")
	ErrTypeAssertion         = errors.New("type assertion failed")
)
