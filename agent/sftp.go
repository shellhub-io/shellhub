package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"syscall"

	"github.com/pkg/sftp"
	"github.com/shellhub-io/shellhub/agent/server/modes/host/command"
)

type pipe struct {
	in  *os.File
	out *os.File
	err *os.File
}

func (p *pipe) Read(data []byte) (int, error) {
	return p.in.Read(data)
}

func (p *pipe) Write(data []byte) (int, error) {
	return p.out.Write(data)
}

func (p *pipe) Close() error {
	os.Exit(0)

	return nil
}

// NewSFTPServer creates a new SFTP server when a new session is created between the agent and the server.
//
// It reads the session user's credentials from HOME, UID, GID and SUPPLEMENTARY_GIDS, as
// Sessioner.SFTP sets them. When the agent runs as root it replaces the supplementary groups, then
// the gid, then the uid, so the server keeps none of the agent's own groups; a missing
// SUPPLEMENTARY_GIDS means no supplementary groups. When the agent does not run as root it switches
// nothing and serves as the agent's own user, as shell and exec sessions do. It serves nothing, and
// reports on stderr, when a credential is missing or malformed, when the chroot to /host fails in
// Docker mode, when it cannot enter HOME, when a credential switch is refused, or when the SFTP
// server cannot start.
func NewSFTPServer(mode command.SFTPServerMode) {
	piped := &pipe{os.Stdin, os.Stdout, os.Stderr}

	if mode == command.SFTPServerModeDocker {
		if err := syscall.Chroot("/host"); err != nil {
			fmt.Fprintln(os.Stderr, err)

			return
		}
	}

	creds, err := sftpCredentialsFromEnv(os.LookupEnv)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)

		return
	}

	if err := syscall.Chdir(creds.home); err != nil {
		fmt.Fprintln(os.Stderr, err)

		return
	}

	if err := processPrivilegeSyscalls.switchTo(creds); err != nil {
		fmt.Fprintln(os.Stderr, err)

		return
	}

	server, err := sftp.NewServer(piped, []sftp.ServerOption{}...)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)

		return
	}

	reportServeError(os.Stderr, server.Serve())

	_ = server.Close()
}

type sftpCredentials struct {
	home   string
	uid    int
	gid    int
	groups []int
}

func sftpCredentialsFromEnv(lookup func(string) (string, bool)) (sftpCredentials, error) {
	home, ok := lookup("HOME")
	if !ok {
		return sftpCredentials{}, errors.New("HOME environment variable not set")
	}

	gid, err := intFromEnv(lookup, "GID")
	if err != nil {
		return sftpCredentials{}, err
	}

	uid, err := intFromEnv(lookup, "UID")
	if err != nil {
		return sftpCredentials{}, err
	}

	encodedGroups, _ := lookup("SUPPLEMENTARY_GIDS")

	groups, err := command.ParseSFTPGroups(encodedGroups)
	if err != nil {
		return sftpCredentials{}, fmt.Errorf("SUPPLEMENTARY_GIDS environment variable is malformed: %w", err)
	}

	return sftpCredentials{home: home, uid: uid, gid: gid, groups: groups}, nil
}

func intFromEnv(lookup func(string) (string, bool), key string) (int, error) {
	value, ok := lookup(key)
	if !ok {
		return 0, fmt.Errorf("%s environment variable not set", key)
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%s environment variable is malformed: %w", key, err)
	}

	return parsed, nil
}

type privilegeSyscalls struct {
	geteuid   func() int
	setgroups func([]int) error
	setgid    func(int) error
	setuid    func(int) error
}

var processPrivilegeSyscalls = privilegeSyscalls{
	geteuid:   os.Geteuid,
	setgroups: syscall.Setgroups,
	setgid:    syscall.Setgid,
	setuid:    syscall.Setuid,
}

func (p privilegeSyscalls) switchTo(creds sftpCredentials) error {
	if p.geteuid() != 0 {
		return nil
	}

	if err := p.setgroups(creds.groups); err != nil {
		return fmt.Errorf("setgroups: %w", err)
	}

	if err := p.setgid(creds.gid); err != nil {
		return fmt.Errorf("setgid: %w", err)
	}

	if err := p.setuid(creds.uid); err != nil {
		return fmt.Errorf("setuid: %w", err)
	}

	return nil
}

func reportServeError(w io.Writer, err error) {
	if err != nil && !errors.Is(err, io.EOF) {
		_, _ = fmt.Fprintln(w, err)
	}
}
