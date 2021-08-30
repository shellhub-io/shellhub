package requests

import (
	"errors"
)

var (
	ErrDeletionReport = errors.New("couldn't report deletion")
	ErrReportUsage    = errors.New("couldn't report the usage")
)
