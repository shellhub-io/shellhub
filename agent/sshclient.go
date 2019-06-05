package main

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type SSHClient struct {
	privateKey string
	host       string
	port       int
	sshPort    int
}

func NewSSHClient(privateKey string, server string, sshPort int) *SSHClient {
	s := &SSHClient{privateKey: privateKey, sshPort: sshPort}

	parts := strings.SplitN(server, ":", 2)

	s.host = parts[0]
	s.port, _ = strconv.Atoi(parts[1])

	return s
}

func (s *SSHClient) connect(msg mqtt.Message) {
	go func() {
		parts := strings.SplitN(string(msg.Payload()), ":", 2)
		port, _ := strconv.Atoi(parts[0])

		args := []string{
			"ssh",
			"-i", s.privateKey,
			"-o", "StrictHostKeyChecking=no",
			"-nNT",
			"-p", fmt.Sprintf("%d", s.port),
			"-R", fmt.Sprintf("%d:localhost:%d", port, s.sshPort),
			fmt.Sprintf("%d:%s@%s", port, parts[1], s.host),
		}

		cmd := exec.Command(args[0], args[1:]...)
		_ = cmd.Start()
	}()
}
