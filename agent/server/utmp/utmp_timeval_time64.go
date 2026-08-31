//go:build arm64
// +build arm64

package utmp

import (
	"github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
)

type TimeVal struct {
	Sec  int64 // Seconds since epoch
	Usec int64 // Microseconds
}

type Utmpx struct {
	Type     int16      // UserProcess or DeadProcess
	Padding1 [2]byte    // Padding to align rest of struct
	Pid      int32      // PID of the ShellHub agent
	Line     [32]byte   // tty associated with the process
	ID       [4]byte    // Index, last 4 characters of Line
	User     [32]byte   // Username
	Host     [256]byte  // Source IP address
	Exit     ExitStatus // Exit status - not used
	Session  int64      // Session ID - not used
	Tv       TimeVal    // Time entry was made
	AddrV6   [4]uint32  // Source IP address. IPv4 in AddrV6[0]
	Reserved [20]byte   // Not used
	Padding2 [4]byte    // Padding to align the next record
}

// This function writes the current time into the utmp record.
func utmpSetTime(u Utmpx) Utmpx {
	a := unix.Timeval{}
	if err := unix.Gettimeofday(&a); err != nil {
		logrus.Warn(err)
	}

	u.Tv.Sec, u.Tv.Usec = a.Sec, int64(a.Usec)

	return u
}
