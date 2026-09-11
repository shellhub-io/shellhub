// Package connectivity reports a server's reachability as transitions rather than as attempts, so
// a component that retries for as long as it is offline logs when it loses the server and when it
// gets it back instead of once per retry.
//
// The attempt number is the transition. A caller whose retry loop already counts attempts, as
// resty's Request.Attempt does, passes the count in to the package-level Lost, Refused and
// Recovered; one that has no counter of its own uses a Tracker, which keeps it.
package connectivity

import (
	"time"

	"github.com/sirupsen/logrus"
)

func level(attempt int) logrus.Level {
	if attempt > 1 {
		return logrus.DebugLevel
	}

	return logrus.WarnLevel
}

// Lost reports one failed attempt to reach the server. The first attempt is logged at warn level
// and every one after it at debug, so an outage costs one line however long it lasts.
func Lost(logger logrus.FieldLogger, attempt int, err error) {
	logger.
		WithError(err).
		WithField("attempt", attempt).
		Log(level(attempt), "Cannot reach the server, retrying until it answers")
}

// Recovered reports the attempt the server finally answered, naming how many it took and how long
// since the first one. Callers report only the attempt that ended a run of failures; one that
// succeeded first time has nothing to say. The wording stays neutral because the run it ends may
// have been a server that could not answer or one that answered and refused, and only the caller
// knows which.
func Recovered(logger logrus.FieldLogger, attempt int, elapsed time.Duration) {
	logger.
		WithFields(logrus.Fields{
			"attempts": attempt,
			"after":    elapsed.String(),
		}).
		Info("Recovered after retrying")
}

const refusalResurfaceEvery = 10

func refusalLevel(attempt int) logrus.Level {
	if attempt%refusalResurfaceEvery == 1 {
		return logrus.WarnLevel
	}

	return logrus.DebugLevel
}

// Refused reports one attempt the server answered by refusing to authorize the device. It reads as
// a distinct condition from an unreachable server because it is: the server is up, and what has to
// change is the namespace or the device limit.
//
// Unlike Lost it does not fall silent after the first attempt. An unreachable server resolves
// itself; a refusal may name a namespace that will never exist, and an operator reading the log
// after the fact needs it to still be saying so. It is raised back to warn every
// refusalResurfaceEvery attempts and logged at debug in between.
func Refused(logger logrus.FieldLogger, attempt int, err error) {
	logger.
		WithError(err).
		WithField("attempt", attempt).
		Log(refusalLevel(attempt), "Cannot authorize the device, retrying until the server accepts it")
}

// Tracker counts consecutive failures for a caller whose retry loop has no attempt counter to pass
// in, and reports them through Lost, Refused and Recovered. It holds no lock: a Tracker belongs to
// the single goroutine running the loop it counts, and two of them covering one server would each
// report the same outage.
type Tracker struct {
	logger  logrus.FieldLogger
	attempt int
	lostAt  time.Time
}

// NewTracker returns a Tracker reporting through logger, with no outage in progress.
func NewTracker(logger logrus.FieldLogger) *Tracker {
	return &Tracker{logger: logger}
}

// Lost counts one failed attempt to reach the server and reports it.
func (t *Tracker) Lost(err error) {
	Lost(t.logger, t.count(), err)
}

// Refused counts one attempt the server answered by refusing the device and reports it.
func (t *Tracker) Refused(err error) {
	Refused(t.logger, t.count(), err)
}

// Recovered reports the attempt that ended a run of failures and clears the count, returning false
// when there was no run to end so the caller can say for itself what a first-time success means.
func (t *Tracker) Recovered() bool {
	if t.attempt == 0 {
		return false
	}

	Recovered(t.logger, t.attempt, time.Since(t.lostAt))
	t.attempt = 0

	return true
}

func (t *Tracker) count() int {
	if t.attempt == 0 {
		t.lostAt = time.Now() //nolint:forbidigo // an elapsed-time measurement: how long the server stays away
	}

	t.attempt++

	return t.attempt
}
