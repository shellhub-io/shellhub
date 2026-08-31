package session

import (
	"errors"
	"net"
	"sync"
)

const (
	maxApprovalWaitsPerSource = 3

	maxApprovalWaitsTotal = 256
)

// ErrApprovalBusy is returned when too many logins are already parked on an
// approval decision.
var ErrApprovalBusy = errors.New("ssh: too many logins waiting for approval")

type approvalLimiter struct {
	perSource int
	total     int

	mu       sync.Mutex
	inFlight map[string]int
	current  int
}

func newApprovalLimiter(perSource, total int) *approvalLimiter {
	return &approvalLimiter{
		perSource: perSource,
		total:     total,
		inFlight:  make(map[string]int),
		current:   0,
	}
}

func (l *approvalLimiter) acquire(addr string) (func(), bool) {
	if isLoopback(addr) {
		return func() {}, true
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if l.current >= l.total || l.inFlight[addr] >= l.perSource {
		return nil, false
	}

	l.inFlight[addr]++
	l.current++

	return sync.OnceFunc(func() { l.release(addr) }), true
}

func (l *approvalLimiter) release(addr string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.inFlight[addr]--; l.inFlight[addr] <= 0 {
		delete(l.inFlight, addr)
	}

	l.current--
}

func isLoopback(addr string) bool {
	ip := net.ParseIP(addr)

	return ip != nil && ip.IsLoopback()
}

var approvals = newApprovalLimiter(maxApprovalWaitsPerSource, maxApprovalWaitsTotal)

func (s *Session) holdApprovalSlot() (func(), error) {
	release, ok := approvals.acquire(s.IPAddress)
	if !ok {
		return nil, ErrApprovalBusy
	}

	return release, nil
}
