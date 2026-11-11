package storetest

import (
	"context"
	"reflect"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/stretchr/testify/assert"
)

type fakeStore interface {
	Scoped(ctx context.Context, sc scope.Scope, uid string) error
	Unscoped(ctx context.Context, uid string) error
	ScopeOutOfPlace(ctx context.Context, uid string, sc scope.Scope) error
	Plumbing() string
}

func TestUnscopedMethods(t *testing.T) {
	cases := []struct {
		description string
		exempt      map[string]string
		expected    []string
	}{
		{
			description: "reports a method taking no scope, or one out of place, unless it is exempt",
			exempt:      map[string]string{"Plumbing": "plumbing"},
			expected: []string{
				"ScopeOutOfPlace takes no scope after its context and is not exempt",
				"Unscoped takes no scope after its context and is not exempt",
			},
		},
		{
			description: "reports an exemption for a method that takes a scope",
			exempt: map[string]string{
				"Plumbing":        "plumbing",
				"Unscoped":        "not converted yet",
				"ScopeOutOfPlace": "not converted yet",
				"Scoped":          "converted already",
			},
			expected: []string{"Scoped takes a scope, so its exemption is stale"},
		},
		{
			description: "reports an exemption naming no method",
			exempt: map[string]string{
				"Plumbing":        "plumbing",
				"Unscoped":        "not converted yet",
				"ScopeOutOfPlace": "not converted yet",
				"Removed":         "deleted since",
			},
			expected: []string{"Removed is exempt but the interface has no such method"},
		},
		{
			description: "reports nothing when every unscoped method is exempt",
			exempt: map[string]string{
				"Plumbing":        "plumbing",
				"Unscoped":        "not converted yet",
				"ScopeOutOfPlace": "not converted yet",
			},
			expected: []string{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			assert.Equal(tt, tc.expected, UnscopedMethods(reflect.TypeFor[fakeStore](), tc.exempt))
		})
	}
}
