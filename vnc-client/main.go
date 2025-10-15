package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

// StartDefault starts the VNC-over-SSH client using defaults:
// - SSH server: localhost:22
// - Local listen port: 5900
func StartDefault() error {
	return Start("localhost:22", 5900)
}

// Start connects to the SSH server at serverAddr and listens on localPort.
// For every incoming TCP connection it opens an SSH channel named "vnc"
// and proxies data bidirectionally.
func Start(serverAddr string, localPort int) error {
	// sshUser := os.Getenv("SSH_USER")
	// if sshUser == "" {
	// 	if u, err := user.Current(); err == nil {
	// 		sshUser = u.Username
	// 	} else {
	// 		sshUser = "root"
	// 	}
	// }

	sshUser := "henry@namespace.60-18-95-75-6a-9d"

	config, err := sshClientConfig(sshUser)
	if err != nil {
		return fmt.Errorf("prepare ssh config: %w", err)
	}

	client, err := ssh.Dial("tcp", serverAddr, config)
	if err != nil {
		return fmt.Errorf("failed to dial ssh server %s: %w", serverAddr, err)
	}
	log.Printf("connected to ssh server %s", serverAddr)

	listenAddr := fmt.Sprintf(":%d", localPort)
	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		client.Close()
		return fmt.Errorf("listen %s: %w", listenAddr, err)
	}
	log.Printf("listening on %s -> ssh channel 'vnc'", listenAddr)

	for {
		localConn, err := ln.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				time.Sleep(100 * time.Millisecond)
				continue
			}
			_ = ln.Close()
			_ = client.Close()
			return fmt.Errorf("accept error: %w", err)
		}

		go handleConn(client, localConn)
	}
}

type Data struct {
	Display string
}

func handleConn(client *ssh.Client, localConn net.Conn) {
	defer localConn.Close()

	d := ssh.Marshal(Data{Display: ":0"})

	ch, reqs, err := client.OpenChannel("vnc", d)
	if err != nil {
		log.Printf("open vnc channel: %v", err)
		return
	}

	defer ch.Close()

	// discard requests from the channel
	go func(in <-chan *ssh.Request) {
		for req := range in {
			// reply false for any global requests to avoid blocking
			if req.WantReply {
				_ = req.Reply(false, nil)
			}
		}
	}(reqs)

	done := make(chan struct{}, 2)

	go func() {
		_, _ = io.Copy(ch, localConn)
		// try to half-close the channel's write side
		_ = ch.CloseWrite()
		done <- struct{}{}
	}()

	go func() {
		_, _ = io.Copy(localConn, ch)
		// try to half-close the local connection's write side when possible
		if tc, ok := localConn.(*net.TCPConn); ok {
			_ = tc.CloseWrite()
		}
		done <- struct{}{}
	}()

	<-done
	<-done
}

// sshClientConfig builds an ssh.ClientConfig using the agent or ~/.ssh/id_rsa
// or SSH_PASSWORD environment variable as fallbacks.
func sshClientConfig(user string) (*ssh.ClientConfig, error) {
	auths := []ssh.AuthMethod{}

	// try ssh-agent
	if sock := os.Getenv("SSH_AUTH_SOCK"); sock != "" {
		if conn, err := net.Dial("unix", sock); err == nil {
			ag := agent.NewClient(conn)
			auths = append(auths, ssh.PublicKeysCallback(ag.Signers))
		}
	}

	// try default key file
	keyPath := filepath.Join(os.Getenv("HOME"), ".ssh", "id_rsa")
	if fi, err := os.Stat(keyPath); err == nil && fi.Mode().IsRegular() {
		if signer, err := signerFromFile(keyPath); err == nil {
			auths = append(auths, ssh.PublicKeys(signer))
		}
	}

	auths = append(auths, ssh.Password(os.Getenv("SSH_PASSWORD")))

	if len(auths) == 0 {
		return nil, fmt.Errorf("no ssh auth method available (set SSH_AUTH_SOCK, provide ~/.ssh/id_rsa or SSH_PASSWORD)")
	}

	cfg := &ssh.ClientConfig{
		User:            user,
		Auth:            auths,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	return cfg, nil
}

func signerFromFile(path string) (ssh.Signer, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ssh.ParsePrivateKey(b)
}

func main() {
	if err := StartDefault(); err != nil {
		log.Fatalf("error: %v", err)
	}
}
