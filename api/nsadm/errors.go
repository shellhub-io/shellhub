package nsadm

import (
	"errors"
)

var (
	ErrUnauthorized      = errors.New("unauthorized")
	ErrUserNotFound      = errors.New("user not found")
	ErrNamespaceNotFound = errors.New("namespace not found")
	ErrDuplicateID       = errors.New("user already member of this namespace")
	ErrConflictName      = errors.New("this name already exists")
	ErrInvalidFormat     = errors.New("invalid name format")
	ErrBadRequest        = errors.New("bad request")
)
