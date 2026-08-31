package utmp

import (
	"bytes"
	"encoding/binary"
	"io"
	"net"
	"os"
	"strings"
	"unsafe"

	"github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
)

type ExitStatus struct {
	ETermination int16 // Process temination status - not used
	EExit        int16 // Process exit status - not used
}

const (
	UtmpxFile   = "/var/run/utmp"
	WtmpxFile   = "/var/log/wtmp"
	UserProcess = 0x7 // Normal process
	DeadProcess = 0x8 // Terminated process
)

// UtmpStartSession This function updates the utmp and wtmp files at the start of a user session.
func UtmpStartSession(line, user, remoteAddr string) Utmpx {
	var u Utmpx

	u.Type = UserProcess
	u.Pid = int32(os.Getpid()) //nolint:gosec
	u = utmpSetTime(u)

	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"ip": remoteAddr,
		}).Warn("wrong remoteAddr format")
	} else {
		ip := net.ParseIP(host)
		if ip4 := ip.To4(); ip4 != nil {
			u.AddrV6[0] = binary.LittleEndian.Uint32(ip4)
		} else {
			u.AddrV6[0] = binary.LittleEndian.Uint32(ip[0:4])
			u.AddrV6[1] = binary.LittleEndian.Uint32(ip[4:8])
			u.AddrV6[2] = binary.LittleEndian.Uint32(ip[8:12])
			u.AddrV6[3] = binary.LittleEndian.Uint32(ip[12:16])
		}
	}

	line = strings.TrimPrefix(line, "/dev/")
	id := line[len(line)-4:]

	_ = copy(u.ID[:], id)
	_ = copy(u.Line[:], line)
	_ = copy(u.User[:], user)
	_ = copy(u.Host[:], host)

	updUtmp(u, id)
	updWtmp(u)

	return u
}

// UtmpEndSession this function updates the utmp and wtmp files at the end of a user session.
func UtmpEndSession(u Utmpx) {
	u.Type = DeadProcess
	u.User = [32]byte{}
	u.Host = [256]byte{}
	u = utmpSetTime(u)

	updUtmp(u, string(u.ID[:]))

	u.ID = [4]byte{}
	u.AddrV6 = [4]uint32{}

	updWtmp(u)
}

func updUtmp(u Utmpx, id string) {
	file, err := os.OpenFile( //nolint:gosec // utmp files conventionally use 0644; tighter permissions break system accounting tools.
		UtmpxFile,
		os.O_RDWR|os.O_CREATE,
		0o644)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"file": UtmpxFile,
			"err":  err,
		}).Warn("Open failed")

		return
	}

	defer file.Close() //nolint:errcheck

	lk := unix.Flock_t{
		Type: int16(unix.F_WRLCK),
		Pid:  int32(os.Getpid()), //nolint:gosec // The maximum value of a pid in Linux and FreeBSD systems fits inside a int32.
	}

	err = unix.FcntlFlock(file.Fd(), unix.F_SETLKW, &lk)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"file": UtmpxFile,
			"err":  err,
		}).Warn("Lock failed")

		return
	}

	var ut Utmpx

	for {
		offset, err := file.Seek(0, io.SeekCurrent)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"file": UtmpxFile,
				"err":  err,
			}).Warn("Null seek failed")

			return
		}

		err = binary.Read(file, binary.LittleEndian, &ut)
		if err != nil {
			break // EOF found: no record with index id
		}

		utID := string(bytes.Trim(ut.ID[:], "\x00"))

		if utID == id {
			if _, err = file.Seek(offset, io.SeekStart); err != nil {
				logrus.WithFields(logrus.Fields{
					"file": UtmpxFile,
					"err":  err,
				}).Warn("Back seek failed")

				return
			}

			break
		}
	}

	if err := binary.Write(file, binary.LittleEndian, &u); err != nil {
		logrus.WithFields(logrus.Fields{
			"file": UtmpxFile,
			"err":  err,
		}).Warn("Write failed utmp")
	}
}

func updWtmp(u Utmpx) {
	file, err := os.OpenFile( //nolint:gosec // wtmp files conventionally use 0644; tighter permissions break system accounting tools.
		WtmpxFile,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0o644)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"file": WtmpxFile,
			"err":  err,
		}).Warn("Open failed")

		return
	}

	defer file.Close() //nolint:errcheck

	lk := unix.Flock_t{
		Type: int16(unix.F_WRLCK),
		Pid:  int32(os.Getpid()), //nolint:gosec // The maximum value of a pid in Linux and FreeBSD systems fits inside a int32.
	}

	err = unix.FcntlFlock(file.Fd(), unix.F_SETLKW, &lk)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"file": WtmpxFile,
			"err":  err,
		}).Warn("Lock failed")
	}

	fileSize, err := file.Seek(0, io.SeekEnd)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"file": WtmpxFile,
			"err":  err,
		}).Warn("Seek to end failed")

		return
	}

	rem := fileSize % int64(unsafe.Sizeof(Utmpx{}))
	if rem != 0 {
		fileSize -= rem
		logrus.WithFields(logrus.Fields{
			"file":     WtmpxFile,
			"filesize": fileSize,
		}).Warn("Database size invalid, truncating")

		if err := file.Truncate(fileSize); err != nil {
			logrus.WithFields(logrus.Fields{
				"file": WtmpxFile,
				"err":  err,
			}).Warn("Database truncate failed")
		}
	}

	if err := binary.Write(file, binary.LittleEndian, &u); err != nil {
		logrus.WithFields(logrus.Fields{
			"file": WtmpxFile,
			"err":  err,
		}).Warn("Write failed on wtmp")

		if err := file.Truncate(fileSize); err != nil {
			logrus.WithFields(logrus.Fields{
				"file": WtmpxFile,
				"err":  err,
			}).Warn("Database truncate failed")
		}
	}
}
