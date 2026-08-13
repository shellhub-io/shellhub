// Package authctx carries what the authentication step already resolved into the request
// context, so the rest of the request can reuse it instead of reading the store again.
package authctx

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
)

type ctxKey string

const deviceLimitKey ctxKey = "namespace-device-limit"

// namespaceDeviceLimit pairs the limit with the namespace it was read for, so a request
// targeting a different namespace cannot be answered from it.
type namespaceDeviceLimit struct {
	tenantID string
	limit    models.NamespaceDeviceLimit
}

// WithNamespaceDeviceLimit returns ctx carrying tenantID's device limit as authentication read it.
func WithNamespaceDeviceLimit(ctx context.Context, tenantID string, limit models.NamespaceDeviceLimit) context.Context {
	return context.WithValue(ctx, deviceLimitKey, namespaceDeviceLimit{tenantID: tenantID, limit: limit})
}

// NamespaceDeviceLimit returns tenantID's device limit as it stood when the request
// authenticated. It is a snapshot, not a live read: a device accepted after authentication is
// not reflected, which matches the handler it replaces — that one re-read the namespace within
// the same request.
//
// It reports false when the request carries no limit, which is the ordinary case for a token
// with no tenant, for API keys, for the admin surface, and for contexts built outside an HTTP
// request; and when the limit belongs to another namespace. Callers must fall back to the store
// rather than treat either as an error.
func NamespaceDeviceLimit(ctx context.Context, tenantID string) (models.NamespaceDeviceLimit, bool) {
	held, ok := ctx.Value(deviceLimitKey).(namespaceDeviceLimit)
	if !ok || held.tenantID != tenantID {
		return models.NamespaceDeviceLimit{}, false
	}

	return held.limit, true
}
