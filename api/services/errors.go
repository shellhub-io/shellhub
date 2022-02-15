package services

import (
	"errors"
)

var (
	ErrNotFound                  = errors.New("not found")
	ErrConflict                  = errors.New("conflict")
	ErrBadRequest                = errors.New("bad request")
	ErrUnauthorized              = errors.New("unauthorized")
	ErrForbidden                 = errors.New("forbidden")
	ErrUserNotFound              = errors.New("user not found")
	ErrNamespaceNotFound         = errors.New("namespace not found")
	ErrNamespaceMemberNotFound   = errors.New("member not found on namespace")
	ErrNamespaceDuplicatedMember = errors.New("namespace already have this member")
	ErrDuplicateID               = errors.New("user already member of this namespace")
	ErrConflictName              = errors.New("this Name already exists")
	ErrInvalidFormat             = errors.New("invalid Name format")
	ErrMaxDeviceCountReached     = errors.New("maximum number of accepted devices reached")
	ErrDuplicatedDeviceName      = errors.New("the Name already exists in the namespace")
	ErrDuplicateFingerprint      = errors.New("this fingerprint already exits")
	ErrReport                    = errors.New("report error")
	ErrMaxTagReached             = errors.New("maximum number of tags reached")
	ErrDeviceNotFound            = errors.New("device not found")
	ErrDuplicateTagName          = errors.New("this tag Name already exists")
	ErrTagNameNotFound           = errors.New("tag Name not found")
	ErrPublicKeyNotFound         = errors.New("public key not found")
	ErrPublicKeyInvalid          = errors.New("public key invalid")
	ErrNoTags                    = errors.New("no tags has found")
	ErrTypeAssertion             = errors.New("type assertion failed")
)
