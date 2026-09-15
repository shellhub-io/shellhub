package gateway

import (
	"net/http"
	"slices"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
)

// Identity is the caller's authenticated identity. It is produced by the
// authentication step and consumed by handlers through the accessors on
// [Context], which read it back from the canonical headers.
type Identity struct {
	ID        string
	Username  string
	TenantID  string
	DeviceUID string
	APIKey    string
	Role      authorizer.Role
	Admin     bool
}

var identityHeaders = []string{
	"X-ID",
	"X-Username",
	"X-Tenant-ID",
	"X-Device-UID",
	"X-API-Key",
	"X-Role",
	"X-Admin",
}

// IdentityHeaders returns the headers [Identity.WriteTo] stamps. A request dispatched internally
// must carry them forward for the caller's identity to survive the hop, and reading them from here
// is what keeps that set from drifting out of step with the write.
func IdentityHeaders() []string {
	return slices.Clone(identityHeaders)
}

// WriteTo stamps the identity onto header, clearing every identity header
// first — including the ones this identity leaves empty.
//
// Clearing unconditionally is the point: a client-supplied X-ID must never
// reach a handler, on an authenticated request or an anonymous one. Requests
// arrive from the edge proxy, which forwards client headers verbatim, so the
// header set is only trustworthy once it has been overwritten here.
func (i *Identity) WriteTo(header http.Header) {
	for _, name := range identityHeaders {
		header.Del(name)
	}

	if i == nil {
		return
	}

	set := func(name, value string) {
		if value != "" {
			header.Set(name, value)
		}
	}

	set("X-ID", i.ID)
	set("X-Username", i.Username)
	set("X-Tenant-ID", i.TenantID)
	set("X-Device-UID", i.DeviceUID)
	set("X-API-Key", i.APIKey)
	set("X-Role", i.Role.String())

	if i.Admin {
		header.Set("X-Admin", "true")
	}
}

// IdentityFrom reads back the identity [Identity.WriteTo] stamped onto header.
//
// It is the read side of that write, and the two must name the same headers. TestIdentityRoundTrip
// is what holds them together; a comment cannot.
func IdentityFrom(header http.Header) Identity {
	return Identity{
		ID:        header.Get("X-ID"),
		Username:  header.Get("X-Username"),
		TenantID:  header.Get("X-Tenant-ID"),
		DeviceUID: header.Get("X-Device-UID"),
		APIKey:    header.Get("X-API-Key"),
		Role:      authorizer.RoleFromString(header.Get("X-Role")),
		Admin:     header.Get("X-Admin") == "true",
	}
}

// Actor returns the identity as the [Actor] a handler receives: who is performing the request,
// without the role and admin flag. Those decide what the caller may do, which the middleware
// answers before a handler runs.
func (i *Identity) Actor() Actor {
	return Actor{ID: i.ID, Username: i.Username, APIKey: i.APIKey, DeviceUID: i.DeviceUID}
}

// WithoutUserScope returns the identity stripped of the acting user's ID and
// namespace scope, keeping the admin flag.
//
// The /admin/api surface depends on this shape: [middleware.Authorize] and
// [middleware.RequiresTenant] treat a missing X-ID alongside X-Admin as "this
// call came from the admin panel rather than from a namespace member", and let
// it past the tenant guard. An admin hitting the regular /api surface keeps
// their user scope and stays subject to that guard.
func (i *Identity) WithoutUserScope() Identity {
	scoped := *i
	scoped.ID = ""
	scoped.TenantID = ""

	return scoped
}
