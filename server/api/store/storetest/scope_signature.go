package storetest

import (
	"context"
	"reflect"
	"slices"

	"github.com/shellhub-io/shellhub/pkg/api/scope"
)

// UnscopedMethods walks the methods of the interface type iface and reports, sorted, every method
// that takes no [scope.Scope] right after its [context.Context] and is not a key of exempt, and
// every key of exempt that names a method taking one or no method of iface at all, so an exemption
// outliving its reason fails as loudly as a missing scope. It panics when iface is not an
// interface type.
func UnscopedMethods(iface reflect.Type, exempt map[string]string) []string {
	complaints := make([]string, 0)

	for method := range iface.Methods() {
		_, exempted := exempt[method.Name]

		switch scoped := TakesScope(method.Type); {
		case !scoped && !exempted:
			complaints = append(complaints, method.Name+" takes no scope after its context and is not exempt")
		case scoped && exempted:
			complaints = append(complaints, method.Name+" takes a scope, so its exemption is stale")
		}
	}

	for name := range exempt {
		if _, ok := iface.MethodByName(name); !ok {
			complaints = append(complaints, name+" is exempt but the interface has no such method")
		}
	}

	slices.Sort(complaints)

	return complaints
}

// TakesScope reports whether the method type method, as an interface's method set lists it with
// no receiver, takes a [context.Context] and then a [scope.Scope] as its first two parameters.
func TakesScope(method reflect.Type) bool {
	return method.NumIn() >= 2 &&
		method.In(0) == reflect.TypeFor[context.Context]() &&
		method.In(1) == reflect.TypeFor[scope.Scope]()
}
