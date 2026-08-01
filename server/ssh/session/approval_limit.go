package session

import (
	"errors"
	"net"
	"sync"
)

const (
	// maxApprovalWaitsPerSource bounds how many logins from one address may sit
	// waiting on a browser approval at the same time. A person approving a login
	// needs one; a handful covers someone opening several terminals at once.
	maxApprovalWaitsPerSource = 3

	// maxApprovalWaitsTotal bounds the whole process, so a botnet spread across
	// many addresses cannot do what one address cannot.
	maxApprovalWaitsTotal = 256
)

// ErrApprovalBusy is returned when too many logins are already parked on an
// approval decision.
var ErrApprovalBusy = errors.New("ssh: too many logins waiting for approval")

// approvalLimiter bounds the logins parked on a browser approval.
//
// Each wait holds a goroutine and an approval row for up to the approval
// window, which is the most expensive thing an unauthenticated-but-key-holding
// client can ask the gateway to do. Counting the waits bounds that directly,
// rather than proxying for it by counting connections: a connection cap keys on
// the wrong thing, since a NAT'd office legitimately opens many.
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

// acquire takes a slot for addr, returning the release func and whether it was
// granted.
//
// Loopback is exempt: the web terminal bridge dials the gateway over it, so
// every browser session in the instance would otherwise share one budget.
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

// holdApprovalSlot reserves this session's place among the logins waiting on an
// approval. The returned func gives it back and is safe to call more than once.
func (s *Session) holdApprovalSlot() (func(), error) {
	release, ok := approvals.acquire(s.IPAddress)
	if !ok {
		return nil, ErrApprovalBusy
	}

	return release, nil
}
