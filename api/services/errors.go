package services

import (
	"errors"
)

var (
	ErrNotFound              = errors.New("not found")
	ErrConflict              = errors.New("conflict")
	ErrBadRequest            = errors.New("bad request")
	ErrUnauthorized          = errors.New("unauthorized")
	ErrForbidden             = errors.New("forbidden")
	ErrUserNotFound          = errors.New("user not found")
	ErrNamespaceNotFound     = errors.New("namespace not found")
	ErrDuplicateID           = errors.New("user already member of this namespace")
	ErrConflictName          = errors.New("this name already exists")
	ErrInvalidFormat         = errors.New("invalid name format")
	ErrMaxDeviceCountReached = errors.New("maximum number of accepted devices reached")
	ErrDuplicatedDeviceName  = errors.New("the name already exists in the namespace")
	ErrDuplicateFingerprint  = errors.New("this fingerprint already exits")
	ErrReport                = errors.New("report error")
	ErrMaxTagReached         = errors.New("maximum number of tags reached")
	ErrDeviceNotFound        = errors.New("device not found")
	ErrDuplicateTagName      = errors.New("this tag name already exists")
	ErrTagNameNotFound       = errors.New("tag name not found")
	ErrNoTags                = errors.New("no tags has found")
)
