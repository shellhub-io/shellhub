package services

import (
	"context"
	"fmt"

	"github.com/shellhub-io/shellhub/api/pkg/responses"
)

// SystemInfoHookFn is called after GetSystemInfo builds its response, allowing
// enterprise/cloud packages to augment the response (e.g. add SAML status).
// Hooks run synchronously; a returned error aborts the request.
type SystemInfoHookFn func(ctx context.Context, info *responses.SystemInfo) error

var systemInfoHooks []SystemInfoHookFn

// OnGetSystemInfo registers a hook that fires after GetSystemInfo builds its
// response. Must be called during package init, before the server starts
// handling requests. Cloud/enterprise packages use this to inject additional
// authentication capabilities (e.g. SAML enabled status) into the /info response.
func OnGetSystemInfo(fn SystemInfoHookFn) {
	if fn == nil {
		panic("services: OnGetSystemInfo called with nil hook")
	}

	systemInfoHooks = append(systemInfoHooks, fn)
}

// fireGetSystemInfo dispatches all registered system info hooks sequentially.
func fireGetSystemInfo(ctx context.Context, info *responses.SystemInfo) error {
	for _, fn := range systemInfoHooks {
		if err := fn(ctx, info); err != nil {
			return fmt.Errorf("system info hook failed: %w", err)
		}
	}

	return nil
}
