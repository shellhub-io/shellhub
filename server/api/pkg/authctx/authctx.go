// Package authctx carries what the authentication step already resolved into the request
// context, so the rest of the request can reuse it instead of reading the store again.
package authctx

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
)

type ctxKey string

const (
	deviceLimitKey ctxKey = "namespace-device-limit"
	userBearerKey  ctxKey = "user-bearer"
)

type userBearer struct {
	userID       string
	version      int
	impersonated bool
}

// WithUserBearer returns ctx carrying what the bearer userID authenticated with says: its token
// version, and whether an instance administrator was issued it to act as the user.
func WithUserBearer(ctx context.Context, userID string, version int, impersonated bool) context.Context {
	return context.WithValue(ctx, userBearerKey, userBearer{userID: userID, version: version, impersonated: impersonated})
}

// UserTokenVersion returns the token version of the bearer userID authenticated with, which
// authentication found equal to the user's own. It reports false when the request did not
// authenticate with a user bearer, which covers API keys, device tokens, anonymous requests and
// contexts built outside an HTTP request, and when the bearer belongs to another user.
func UserTokenVersion(ctx context.Context, userID string) (int, bool) {
	held, ok := ctx.Value(userBearerKey).(userBearer)
	if !ok || held.userID != userID {
		return 0, false
	}

	return held.version, true
}

// UserImpersonated reports whether the bearer userID authenticated with was issued to an
// instance administrator acting as the user. It reports false for every request
// [UserTokenVersion] reports false for.
func UserImpersonated(ctx context.Context, userID string) bool {
	held, ok := ctx.Value(userBearerKey).(userBearer)

	return ok && held.userID == userID && held.impersonated
}

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
