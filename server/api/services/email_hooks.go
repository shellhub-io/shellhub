package services

import "context"

// EmailCheckHookFn is called before an address the user typed is written to their account, on
// open sign-up and on a change of email or recovery email. It returns the reason it refuses the
// address, which becomes the field's message in the 400, or "" to accept it. It runs on the
// request's context, so a hook that does I/O must bound it itself; it has no way to fail other
// than accepting or refusing.
type EmailCheckHookFn func(ctx context.Context, email string) string

var emailCheckHooks []EmailCheckHookFn

// OnEmailCheck registers a hook that may refuse an address before it is written. It must be called
// during package init, before the server starts handling requests. With no hook registered every
// address is accepted. It panics when fn is nil.
func OnEmailCheck(fn EmailCheckHookFn) {
	if fn == nil {
		panic("services: OnEmailCheck called with nil hook")
	}

	emailCheckHooks = append(emailCheckHooks, fn)
}

func checkEmail(ctx context.Context, email string) string {
	for _, fn := range emailCheckHooks {
		if reason := fn(ctx, email); reason != "" {
			return reason
		}
	}

	return ""
}
