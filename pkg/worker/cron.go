package worker

import (
	"context"

	"github.com/adhocore/gronx"
)

// CronSpec is a cron expression in the five-field form. Validate before use: an invalid spec would
// otherwise register a job that never fires.
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

// CronHandler runs one scheduled execution. It takes no payload — a cron job's only input is the
// fact that it is time.
type CronHandler func(ctx context.Context) error

// Cronjob binds a handler to a schedule. Unique is what stops a long run from overlapping the next
// tick, which matters for jobs that sweep shared state.
type Cronjob struct {
	// Identifier is a UUID for the cron job, used internally to register the task with the
	// scheduler.
	Identifier string
	// Spec is the cron expression that defines the schedule for the cron job.
	Spec CronSpec
	// Handler is the callback function that will be executed when the cron specification is met.
	Handler CronHandler
	// Unique defines whether the task cannot be perfomed concurrently.
	Unique bool
}

// CronjobOption configures a cron job at registration.
type CronjobOption func(c *Cronjob)
