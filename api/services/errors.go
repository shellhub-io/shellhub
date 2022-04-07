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
	// Data is a key-value map of the invalid fields. key must be the field name what is invalid and value must be the
	// value of the field.
	Data map[string]interface{}
}

var (
	ErrReport                    = errors.New("report error", ErrLayer, ErrCodeInvalid)
	ErrNotFound                  = errors.New("not found", ErrLayer, ErrCodeNotFound)
	ErrBadRequest                = errors.New("bad request", ErrLayer, ErrCodeInvalid)
	ErrUnauthorized              = errors.New("unauthorized", ErrLayer, ErrCodeInvalid)
	ErrForbidden                 = errors.New("forbidden", ErrLayer, ErrCodeNotFound)
	ErrUserNotFound              = errors.New("user not found", ErrLayer, ErrCodeNotFound)
	ErrUserInvalid               = errors.New("user invalid", ErrLayer, ErrCodeInvalid)
	ErrUserDuplicated            = errors.New("user duplicated", ErrLayer, ErrCodeDuplicated)
	ErrUserPasswordInvalid       = errors.New("user password invalid", ErrLayer, ErrCodeInvalid)
	ErrUserPasswordDuplicated    = errors.New("user password is equal to new password", ErrLayer, ErrCodeDuplicated)
	ErrUserPasswordNotMatch      = errors.New("user password does not match to the current password", ErrLayer, ErrCodeInvalid)
	ErrNamespaceNotFound         = errors.New("namespace not found", ErrLayer, ErrCodeNotFound)
	ErrNamespaceMemberNotFound   = errors.New("member not found", ErrLayer, ErrCodeNotFound)
	ErrNamespaceDuplicatedMember = errors.New("member duplicated", ErrLayer, ErrCodeDuplicated)
	ErrMaxTagReached             = errors.New("tag limit reached", ErrLayer, ErrCodeLimit)
	ErrDuplicateTagName          = errors.New("tag duplicated", ErrLayer, ErrCodeDuplicated)
	ErrTagNameNotFound           = errors.New("tag not found", ErrLayer, ErrCodeNotFound)
	ErrTagInvalid                = errors.New("tag invalid", ErrLayer, ErrCodeInvalid)
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

// NewErrNotFound returns an error with the ErrDataNotFound and wrap an error.
func NewErrNotFound(err error, id string, next error) error {
	return errors.Wrap(errors.WithData(err, ErrDataNotFound{ID: id}), next)
}

// NewErrInvalid returns an error with the ErrDataInvalid and wrap an error.
func NewErrInvalid(err error, data map[string]interface{}, next error) error {
	return errors.Wrap(errors.WithData(err, ErrDataInvalid{Data: data}), next)
}

// NewErrDuplicated returns an error with the ErrDataDuplicated and wrap an error.
func NewErrDuplicated(err error, values []string, next error) error {
	return errors.Wrap(errors.WithData(err, ErrDataDuplicated{Values: values}), next)
}

// NewErrNamespaceNotFound returns an error when the namespace is not found.
func NewErrNamespaceNotFound(id string, next error) error {
	return NewErrNotFound(ErrNamespaceNotFound, id, next)
}

// NewErrTagInvalid returns an error when the tag is invalid.
func NewErrTagInvalid(tag string, next error) error {
	return NewErrInvalid(ErrTagInvalid, map[string]interface{}{"name": tag}, next)
}

// NewErrTagEmpty returns an error when the none tag is found.
func NewErrTagEmpty(tenant string, next error) error {
	return NewErrNotFound(ErrNoTags, tenant, next)
}

// NewErrTagNotFound returns an error when the tag is not found.
func NewErrTagNotFound(tag string, next error) error {
	return NewErrNotFound(ErrTagNameNotFound, tag, next)
}

// NewErrTagDuplicated returns an error when the tag is duplicated.
func NewErrTagDuplicated(tag string, next error) error {
	return NewErrDuplicated(ErrDuplicateTagName, []string{tag}, next)
}

// NewErrUserNotFound returns an error when the user is not found.
func NewErrUserNotFound(id string, next error) error {
	return NewErrNotFound(ErrUserNotFound, id, next)
}

// NewErrUserInvalid returns an error when the user is invalid.
func NewErrUserInvalid(data map[string]interface{}, next error) error {
	return NewErrInvalid(ErrUserInvalid, data, next)
}

// NewErrUserDuplicated returns an error when the user is duplicated.
func NewErrUserDuplicated(values []string, next error) error {
	return NewErrDuplicated(ErrUserDuplicated, values, next)
}

// NewErrUserPasswordInvalid returns an error when the user's password is invalid.
func NewErrUserPasswordInvalid(next error) error {
	return NewErrInvalid(ErrUserPasswordInvalid, nil, next)
}

// NewErrUserPasswordDuplicated returns an error when the user's current password is equal to new password.
func NewErrUserPasswordDuplicated(next error) error {
	return NewErrDuplicated(ErrUserPasswordDuplicated, nil, next)
}

// NewErrUserPasswordNotMatch returns an error when the user's password doesn't match with the current password.
func NewErrUserPasswordNotMatch(next error) error {
	return NewErrInvalid(ErrUserPasswordNotMatch, nil, next)
}
