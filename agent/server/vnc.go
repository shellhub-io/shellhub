package server

import (
	"fmt"
	"log"

	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/pkg/vnc"
	gossh "golang.org/x/crypto/ssh"
)

type Data struct {
	Display string
}

func VNC(srv *gliderssh.Server, conn *gossh.ServerConn, newChan gossh.NewChannel, ctx gliderssh.Context) {
	fmt.Println("VNC channel requested")
	defer fmt.Println("VNC channel closed")

	d := newChan.ExtraData()

	var data Data
	if err := gossh.Unmarshal(d, &data); err != nil {
		log.Printf("Failed to unmarshal VNC channel data: %v", err)
		newChan.Reject(gossh.Prohibited, "invalid channel data") //nolint:errcheck

		return
	}

	ch, reqs, err := newChan.Accept()
	if err != nil {
		log.Printf("Could not accept channel: %v", err)
		newChan.Reject(gossh.ConnectionFailed, "could not accept channel") //nolint:errcheck

		return
	}

	// Discard all global out-of-band Requests
	go gossh.DiscardRequests(reqs)

	x11Display, err := vnc.NewX11Display(data.Display)
	if err != nil {
		log.Printf("Failed to create X11 display: %v", err)
		newChan.Reject(gossh.ConnectionFailed, "failed to create X11 display") //nolint:errcheck

		return
	}

	// Start VNC server
	server, err := vnc.NewVNCServer(&vnc.Config{
		Name:  "ShellHub VNC Server",
		Auths: []vnc.Auth{vnc.NewNoAuth()},
	})
	if err != nil {
		log.Printf("Failed to create VNC server: %v", err)
		newChan.Reject(gossh.ConnectionFailed, "failed to create VNC server") //nolint:errcheck

		return
	}

	server.Handle(ch, x11Display)
}
