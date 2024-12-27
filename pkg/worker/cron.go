package worker

import (
	"context"

	"github.com/adhocore/gronx"
)

type CronSpec string

func (cs CronSpec) String() string {
	return string(cs)
}

// _gron is only used to validate cron expressions. It's initialized outside of the
// checker to avoid multiples allocs of the struct.
var _gron = gronx.New()

// Validate reports whether the pattern is valid or not.
func (cs CronSpec) Validate() bool {
	return _gron.IsValid(cs.String())
}

// MustValidate is similar to [CronSpec.Validate] but panics when invalid.
func (cs CronSpec) MustValidate() {
	if !cs.Validate() {
		panic("invalid cron specification: " + cs)
	}
}

type CronHandler func(ctx context.Context) error

type Cronjob struct {
	Handler    CronHandler
	Identifier string
	Spec       CronSpec
	Unique     bool
}

type CronjobOption func(c *Cronjob)
