package session

import (
	"errors"
	"net"

	gossh "golang.org/x/crypto/ssh"
)

func deviceHostKey(device gossh.PublicKey) gossh.HostKeyCallback {
	pinned := gossh.FixedHostKey(device)

	return func(hostname string, remote net.Addr, key gossh.PublicKey) error {
		if err := pinned(hostname, remote, key); err != nil {
			return errors.Join(ErrHostKey, err)
		}

		return nil
	}
}
