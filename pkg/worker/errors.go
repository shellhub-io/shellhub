package worker

import "errors"

// The errors this package returns. They are compared with errors.Is, so a caller can tell a
// misconfigured worker (an invalid pattern or spec, caught at registration) from a runtime failure.
var (
	ErrHandleCronFailed   = errors.New("failed to handle cron")
	ErrServerStartFailed  = errors.New("failed to start the worker server")
	ErrClientStartFailed  = errors.New("failed to start the worker client")
	ErrTaskPatternInvalid = errors.New("task pattern is invalid")
	ErrCronSpecInvalid    = errors.New("cron specification is invalid")
	ErrSubmitFailed       = errors.New("failed to submit the payload")
)
