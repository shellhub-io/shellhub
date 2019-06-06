package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"golang.org/x/crypto/ssh"
	"golang.org/x/net/websocket"
)

func copyWorker(dst io.Writer, src io.Reader, doneCh chan<- bool) {
	io.Copy(dst, src)
	doneCh <- true
}

func relayHandler(ws *websocket.Conn) {
	user := ws.Request().URL.Query().Get("user")
	passwd := ws.Request().URL.Query().Get("passwd")
	cols, _ := strconv.Atoi(ws.Request().URL.Query().Get("cols"))
	rows, _ := strconv.Atoi(ws.Request().URL.Query().Get("rows"))

	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(passwd),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	client, err := ssh.Dial("tcp", "ssh:2222", config)
	if err != nil {
		fmt.Println(err)
		ws.Close()
		return
	}

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	session, err := client.NewSession()
	if err != nil {
		session.Close()
		ws.Close()
		return
	}

	sshOut, err := session.StdoutPipe()
	if err != nil {
		session.Close()
		ws.Close()
		return
	}

	sshIn, err := session.StdinPipe()
	if err != nil {
		session.Close()
		ws.Close()
		return
	}

	if err := session.RequestPty("xterm", rows, cols, modes); err != nil {
		session.Close()
		ws.Close()
		return
	}
	if err := session.Shell(); err != nil {
		session.Close()
		ws.Close()
		return
	}

	doneCh := make(chan bool)

	go copyWorker(sshIn, ws, doneCh)
	go copyWorker(ws, sshOut, doneCh)

	<-doneCh

	client.Close()
	ws.Close()

	<-doneCh
}

func main() {
	http.Handle("/ssh", websocket.Handler(relayHandler))

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
