package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"strconv"
	"strings"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
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

		key, err := ioutil.ReadFile(s.privateKey)
		if err != nil {
			logrus.Error(err)
			return
		}

		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			logrus.Error(err)
			return
		}

		sshConfig := &ssh.ClientConfig{
			User:            fmt.Sprintf("%d:%s", port, parts[1]),
			Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}

		serverConn, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", s.host, s.port), sshConfig)
		if err != nil {
			logrus.Error(err)
			return
		}

		listener, err := serverConn.Listen("tcp", fmt.Sprintf("localhost:%d", port))
		if err != nil {
			logrus.Error(err)
			return
		}

		defer listener.Close()

		local, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", s.sshPort))
		if err != nil {
			logrus.Error(err)
			return
		}

		client, err := listener.Accept()
		if err != nil {
			logrus.Error(err)
			return
		}

		handleClient(client, local)
	}()
}

func handleClient(client net.Conn, remote net.Conn) {
	defer client.Close()
	chDone := make(chan bool)

	// Start remote -> local data transfer
	go func() {
		_, err := io.Copy(client, remote)
		if err != nil {
			log.Println(fmt.Sprintf("error while copy remote->local: %s", err))
		}
		chDone <- true
	}()

	// Start local -> remote data transfer
	go func() {
		_, err := io.Copy(remote, client)
		if err != nil {
			log.Println(fmt.Sprintf("error while copy local->remote: %s", err))
		}
		chDone <- true
	}()

	<-chDone
}
