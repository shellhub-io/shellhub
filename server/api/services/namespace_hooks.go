package services

import (
	"context"
	"fmt"

	"github.com/shellhub-io/shellhub/pkg/models"
)

// NamespaceDeleteHookFn is called when a namespace is about to be deleted.
// The hook receives the namespace that will be deleted. Hooks run before the
// actual deletion, so a returned error will abort the entire operation.
type NamespaceDeleteHookFn func(ctx context.Context, namespace *models.Namespace) error

var namespaceDeleteHooks []NamespaceDeleteHookFn

// OnNamespaceDelete registers a hook that fires when a namespace is deleted.
// It must be called during package init, before the server starts handling
// requests. Cloud packages use this to clean up cloud-only resources
// (firewall rules, tunnels, recorded sessions, etc.).
func OnNamespaceDelete(fn NamespaceDeleteHookFn) {
	if fn == nil {
		panic("services: OnNamespaceDelete called with nil hook")
	}

	namespaceDeleteHooks = append(namespaceDeleteHooks, fn)
}

// fireNamespaceDelete dispatches all registered delete hooks sequentially.
// The first error aborts execution (the caller rolls back the operation).
func fireNamespaceDelete(ctx context.Context, ns *models.Namespace) error {
	for _, fn := range namespaceDeleteHooks {
		if err := fn(ctx, ns); err != nil {
			return fmt.Errorf("namespace delete hook failed: %w", err)
		}
	}

	return nil
}
