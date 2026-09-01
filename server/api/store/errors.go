package store

import (
	stderrors "errors"

	"github.com/shellhub-io/shellhub/pkg/errors"
)

// ErrLayer is an error level. Each error defined at this level, is container to it.
// ErrLayer is the errors' level for store's error.
const ErrLayer = "store"

// The store's failure codes. A service maps them to its own errors, so no HTTP status is
// decided at this layer.
const (
	ErrCodeNoDocument = iota + 1
	ErrCodeDuplicated
	ErrCodeInvalid
	ErrCodeInternal
	ErrCodeConstraint
)

// The store's failures, each carrying the layer and code above.
var (
	ErrDuplicate        = errors.New("document duplicate", ErrLayer, ErrCodeDuplicated)
	ErrNoDocuments      = errors.New("no documents", ErrLayer, ErrCodeNoDocument)
	ErrInvalidHex       = errors.New("the provided hex string is not a valid ObjectID", ErrLayer, ErrCodeInvalid)
	ErrResolverNotFound = errors.New("resolver not found", ErrLayer, ErrCodeInvalid)
	ErrInternal         = errors.New("internal store error", ErrLayer, ErrCodeInternal)
	// ErrNamespaceInstanceProtected is returned when deleting the namespace bound to the
	// instance (systems.instance_tenant_id) is refused by the FK's ON DELETE RESTRICT.
	ErrNamespaceInstanceProtected = errors.New("namespace is bound to the instance", ErrLayer, ErrCodeConstraint)
	// ErrNamespaceSingle is returned when creating an additional namespace on an instance already
	// bound to one (systems.instance_tenant_id set — Community). Enterprise/Cloud never bind, so
	// this is Community-specific and distinct from a plain duplicate-name conflict.
	ErrNamespaceSingle = errors.New("instance does not support multi-tenancy", ErrLayer, ErrCodeConstraint)
	// ErrInvalidScope is returned when a namespace-bound operation is given a scope that was never
	// constructed. It catches a zero-value [scope.Scope] reaching the store, which would otherwise
	// read as neither bounded nor deliberately unbounded.
	ErrInvalidScope = errors.New("namespace scope is invalid", ErrLayer, ErrCodeInvalid)
)

// DuplicateFieldError carries the name of the field that caused a duplicate-key violation.
// It is a plain Go error type (not a pkg/errors.Error) so that echo's error chain never
// matches it directly; callers use DuplicatedField to extract the field name.
type DuplicateFieldError struct {
	Field string
}

func (e DuplicateFieldError) Error() string {
	return "duplicate field: " + e.Field
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
