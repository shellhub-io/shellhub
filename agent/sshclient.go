package main

import (
	"strings"
)

type SSHClient struct {
	privateKey string
	host       string
	Sessions   []string
}

func NewSSHClient(privateKey string, server string) *SSHClient {
	s := &SSHClient{privateKey: privateKey}

	parts := strings.SplitN(server, ":", 2)

	s.host = parts[0]

	return s
}

func (s *SSHClient) close(id string) {
	for i, v := range s.Sessions {
		if v == id {
			s.Sessions[i] = s.Sessions[len(s.Sessions)-1]
			s.Sessions = s.Sessions[:len(s.Sessions)-1]
		}
	}
}
