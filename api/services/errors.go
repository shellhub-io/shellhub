package services

import (
	"errors"
)

var (
	ErrConflict              = errors.New("conflict")
	ErrBadRequest            = errors.New("bad request")
	ErrUnauthorized          = errors.New("unauthorized")
	ErrUserNotFound          = errors.New("user not found")
	ErrNamespaceNotFound     = errors.New("namespace not found")
	ErrDuplicateID           = errors.New("user already member of this namespace")
	ErrConflictName          = errors.New("this name already exists")
	ErrInvalidFormat         = errors.New("invalid name format")
	ErrMaxDeviceCountReached = errors.New("maximum number of accepted devices reached")
	ErrDuplicatedDeviceName  = errors.New("the name already exists in the namespace")
	ErrDuplicateFingerprint  = errors.New("this fingerprint already exits")
)
