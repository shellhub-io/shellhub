//go:build !arm64

package utmp

import (
	"github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
)

// TimeVal mirrors the kernel's struct timeval for this architecture's word size.
type TimeVal struct {
	Sec  int32 // Seconds since epoch
	Usec int32 // Microseconds
}

// Utmpx is one login accounting record, laid out to match utmp(5) byte for byte so that who,
// w and last can read what the agent writes.
type Utmpx struct {
	Type     int16      // UserProcess or DeadProcess
	Padding  [2]byte    // Padding to align rest of struct
	Pid      int32      // PID of the ShellHub agent
	Line     [32]byte   // tty associated with the process
	ID       [4]byte    // Index, last 4 characters of Line
	User     [32]byte   // Username
	Host     [256]byte  // Source IP address
	Exit     ExitStatus // Exit status - not used
	Session  int32      // Session ID - not used
	Tv       TimeVal    // Time entry was made
	AddrV6   [4]uint32  // Source IP address. IPv4 in AddrV6[0]
	Reserved [20]byte   // Not used
}

func utmpSetTime(u Utmpx) Utmpx {
	a := unix.Timeval{}
	if err := unix.Gettimeofday(&a); err != nil {
		logrus.Warn(err)
	}

	u.Tv.Sec, u.Tv.Usec = int32(a.Sec), int32(a.Usec) //nolint:gosec

	return u
}
