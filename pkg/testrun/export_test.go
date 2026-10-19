package testrun

import "time"

var LeaseAlive = leaseAlive

func SetProcessGraceOf(run string, grace time.Duration) (restore func()) {
	processGraces[run] = grace

	return func() { delete(processGraces, run) }
}
