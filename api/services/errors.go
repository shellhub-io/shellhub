package services

import (
	"github.com/shellhub-io/shellhub/pkg/errors"
)

// ErrLayer is an error level. Each error defined at this level, is container to it.
// ErrLayer is the errors' level for service's error.
const ErrLayer = "service"

const (
	// ErrCodeNotFound is the error code for when a resource is not found.
	ErrCodeNotFound = iota + 1
	// ErrCodeDuplicated is the error code for when a resource is duplicated.
	ErrCodeDuplicated
	// ErrCodeLimit is the error code for when a resource is reached the limit.
	ErrCodeLimit
	// ErrCodeInvalid is the error code for when a resource is invalid.
	ErrCodeInvalid
	// ErrCodePayment is the error code for when a resource required payment.
	ErrCodePayment
	// ErrCodeStore is the error code for when the store function fails. The store function is responsible for execute
	// the main service action.
	ErrCodeStore
)

// ErrDataNotFound structure should be used to add errors.Data to an error when the resource is not found.
type ErrDataNotFound struct {
	// ID is the identifier of the resource.
	ID string
}

// ErrDataDuplicated structure should be used to add errors.Data to an error when the resource is duplicated.
type ErrDataDuplicated struct {
	// Values is used to identify the duplicated resource.
	Values []string
}

// ErrDataLimit structure should be used to add errors.Data to an error when the resource is reached the limit.
type ErrDataLimit struct {
	// Limit is the max number of resources.
	Limit int
}

// ErrDataInvalid structure should be used to add errors.Data to an error when the resource is invalid.
type ErrDataInvalid struct {
	// Values is used to identify the invalid resource.
	Values []string
}

var (
	ErrReport                    = errors.New("report error", ErrLayer, ErrCodeInvalid)
	ErrNotFound                  = errors.New("not found", ErrLayer, ErrCodeNotFound)
	ErrConflict                  = errors.New("conflict", ErrLayer, ErrCodeDuplicated)
	ErrBadRequest                = errors.New("bad request", ErrLayer, ErrCodeInvalid)
	ErrUnauthorized              = errors.New("unauthorized", ErrLayer, ErrCodeInvalid)
	ErrForbidden                 = errors.New("forbidden", ErrLayer, ErrCodeNotFound)
	ErrUserNotFound              = errors.New("user not found", ErrLayer, ErrCodeNotFound)
	ErrNamespaceNotFound         = errors.New("namespace not found", ErrLayer, ErrCodeNotFound)
	ErrNamespaceMemberNotFound   = errors.New("member not found", ErrLayer, ErrCodeNotFound)
	ErrNamespaceDuplicatedMember = errors.New("member duplicated", ErrLayer, ErrCodeDuplicated)
	ErrMaxTagReached             = errors.New("tag limit reached", ErrLayer, ErrCodeLimit)
	ErrDuplicateTagName          = errors.New("tag duplicated", ErrLayer, ErrCodeDuplicated)
	ErrTagNameNotFound           = errors.New("tag not found", ErrLayer, ErrCodeNotFound)
	ErrNoTags                    = errors.New("no tags has found", ErrLayer, ErrCodeNotFound)
	ErrConflictName              = errors.New("name duplicated", ErrLayer, ErrCodeDuplicated)
	ErrInvalidFormat             = errors.New("invalid format", ErrLayer, ErrCodeInvalid)
	ErrDeviceNotFound            = errors.New("device not found", ErrLayer, ErrCodeNotFound)
	ErrMaxDeviceCountReached     = errors.New("maximum number of accepted devices reached", ErrLayer, ErrCodeLimit)
	ErrDuplicatedDeviceName      = errors.New("device name duplicated", ErrLayer, ErrCodeDuplicated)
	ErrDuplicateFingerprint      = errors.New("fingerprint duplicated", ErrLayer, ErrCodeDuplicated)
	ErrPublicKeyNotFound         = errors.New("public key not found", ErrLayer, ErrCodeNotFound)
	ErrPublicKeyInvalid          = errors.New("public key invalid", ErrLayer, ErrCodeInvalid)
	ErrTypeAssertion             = errors.New("type assertion failed", ErrLayer, ErrCodeInvalid)
)
