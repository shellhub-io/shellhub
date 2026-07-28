// Package scope makes the namespace an operation is bounded to a first-class value.
//
// Every namespace-bound operation takes a [Scope] as a required argument, so forgetting to bound a
// query is a compile error rather than something review has to notice. Crossing namespaces is a
// positive act: [NewUnbounded] cannot be called without a reason explaining why it is safe, and a
// grep for it lists every such site in one pass.
//
// Not to be confused with the `scope` database column, which records a namespace's type.
package scope

import (
	"fmt"

	"github.com/shellhub-io/shellhub/pkg/errors"
)

// ErrLayer is the layer that scope errors are reported from.
const ErrLayer = "scope"

// ErrCodeEmptyTenantID is the code reported when a bounded scope is built without a tenant.
const ErrCodeEmptyTenantID = 1

// ErrEmptyTenantID is returned by [NewBounded] when the tenant ID is empty. Unlike the neighbouring
// value types, an invalid input is not normalised to a sentinel: normalising an empty tenant would
// reintroduce the fail-open unbounded query this package exists to remove.
var ErrEmptyTenantID = errors.New("cannot bound a scope to an empty tenant ID", ErrLayer, ErrCodeEmptyTenantID)

// Kind distinguishes the two ways a [Scope] can be constructed. The zero value is [KindInvalid], so
// a [Scope] that was never constructed is neither bounded nor unbounded and stores reject it.
type Kind uint8

const (
	KindInvalid Kind = iota
	KindBounded
	KindUnbounded
)

// Scope is the namespace an operation is bounded to, or an explicit statement that it is not bounded
// at all. Construct it with [NewBounded] or [NewUnbounded]; the zero value is invalid.
type Scope struct {
	kind     Kind
	tenantID string
	reason   string
}

// NewBounded returns a [Scope] bounded to tenantID. It returns [ErrEmptyTenantID] when tenantID is
// empty, so a missing request header becomes an error at the edge instead of an unbounded query deep
// in the store.
func NewBounded(tenantID string) (Scope, error) {
	if tenantID == "" {
		return Scope{}, ErrEmptyTenantID
	}

	return Scope{kind: KindBounded, tenantID: tenantID}, nil
}

// MustBounded is [NewBounded] for tenant IDs that cannot be empty by construction, such as one read
// back from a model the store just returned. It panics on an empty tenant ID.
func MustBounded(tenantID string) Scope {
	sc, err := NewBounded(tenantID)
	if err != nil {
		panic(err)
	}

	return sc
}

// NewUnbounded returns a [Scope] that deliberately spans every namespace. The reason is what makes
// the crossing auditable, so it should say why crossing is safe here, not merely that it happens.
func NewUnbounded(reason string) Scope {
	return Scope{kind: KindUnbounded, reason: reason}
}

// Kind reports how the scope was constructed.
func (s Scope) Kind() Kind {
	return s.kind
}

// IsBounded reports whether the scope is bounded to a namespace.
func (s Scope) IsBounded() bool {
	return s.kind == KindBounded
}

// IsValid reports whether the scope was constructed at all.
func (s Scope) IsValid() bool {
	return s.kind == KindBounded || s.kind == KindUnbounded
}

// TenantID returns the namespace the scope is bounded to, or an empty string when it is not bounded.
func (s Scope) TenantID() string {
	return s.tenantID
}

// Reason returns why the scope deliberately spans namespaces, or an empty string when it is bounded.
func (s Scope) Reason() string {
	return s.reason
}

func (s Scope) String() string {
	switch s.kind {
	case KindBounded:
		return fmt.Sprintf("scope(bounded:%s)", s.tenantID)
	case KindUnbounded:
		return fmt.Sprintf("scope(unbounded:%s)", s.reason)
	default:
		return "scope(invalid)"
	}
}
