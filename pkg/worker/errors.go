package worker

import "errors"

var (
	ErrHandleCronFailed   = errors.New("failed to handle cron")
	ErrServerStartFailed  = errors.New("failed to start the worker server")
	ErrClientStartFailed  = errors.New("failed to start the worker client")
	ErrTaskPatternInvalid = errors.New("task pattern is invalid")
	ErrCronSpecInvalid    = errors.New("cron specification is invalid")
	ErrSubmitFailed       = errors.New("failed to submit the payload")
)
