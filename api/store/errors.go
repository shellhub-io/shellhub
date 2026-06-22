package store

import (
	stderrors "errors"
	"fmt"

	"github.com/shellhub-io/shellhub/pkg/errors"
)

// ErrLayer is an error level. Each error defined at this level, is container to it.
// ErrLayer is the errors' level for store's error.
const ErrLayer = "store"

const (
	ErrCodeNoDocument = iota + 1
	ErrCodeDuplicated
	ErrCodeInvalid
	ErrCodeInternal
)

var (
	ErrDuplicate        = errors.New("document duplicate", ErrLayer, ErrCodeDuplicated)
	ErrNoDocuments      = errors.New("no documents", ErrLayer, ErrCodeNoDocument)
	ErrInvalidHex       = errors.New("the provided hex string is not a valid ObjectID", ErrLayer, ErrCodeInvalid)
	ErrResolverNotFound = errors.New("resolver not found", ErrLayer, ErrCodeInvalid)
	ErrInternal         = errors.New("internal store error", ErrLayer, ErrCodeInternal)
)

// DuplicateFieldError carries the name of the field that caused a duplicate-key violation.
// It is a plain Go error type (not a pkg/errors.Error) so that echo's error chain never
// matches it directly; callers use DuplicatedField to extract the field name.
type DuplicateFieldError struct {
	Field string
}

func (e DuplicateFieldError) Error() string {
	return fmt.Sprintf("duplicate field: %s", e.Field)
}

// DuplicatedField extracts the field name from a DuplicateFieldError wrapped inside err.
// It returns ("", false) when no DuplicateFieldError is present or when Field is empty.
func DuplicatedField(err error) (string, bool) {
	var df DuplicateFieldError
	if stderrors.As(err, &df) && df.Field != "" {
		return df.Field, true
	}

	return "", false
}
