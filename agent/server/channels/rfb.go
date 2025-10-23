package channels

import (
	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/pkg/rfb"
	log "github.com/sirupsen/logrus"
	gossh "golang.org/x/crypto/ssh"
)

func RFBHandler(srv *gliderssh.Server, conn *gossh.ServerConn, newChan gossh.NewChannel, ctx gliderssh.Context) {
	logger := log.WithFields(log.Fields{
		"user":    ctx.User(),
		"channel": newChan.ChannelType(),
		"handler": "RFBHandler",
	})

	logger.Info("New RFB channel request received")
	defer logger.Info("RFB channel handler finished processing")

	d := newChan.ExtraData()

	var data Data
	if err := gossh.Unmarshal(d, &data); err != nil {
		logger.WithError(err).Error("Failed to unmarshal RFB channel data")

		newChan.Reject(gossh.Prohibited, "invalid channel data") //nolint:errcheck

		return
	}

	ch, reqs, err := newChan.Accept()
	if err != nil {
		logger.WithError(err).Error("Failed to accept RFB channel")

		newChan.Reject(gossh.ConnectionFailed, "could not accept channel") //nolint:errcheck

		return
	}

	// NOTE: Discard all global out-of-band Requests.
	go gossh.DiscardRequests(reqs)

	x11Display, err := rfb.NewX11Display(data.Display)
	if err != nil {
		logger.WithError(err).Error("Failed to create X11 display")

		newChan.Reject(gossh.ConnectionFailed, "failed to create X11 display") //nolint:errcheck

		return
	}

	// TODO: Verify which options are necessary.
	server, err := rfb.NewRFBServer(&rfb.Config{
		Name:  "ShellHub RFB Server",
		Auths: []rfb.Auth{rfb.NewNoAuth()},
	})
	if err != nil {
		logger.WithError(err).Error("Failed to create RFB server")

		newChan.Reject(gossh.ConnectionFailed, "failed to create RFB server") //nolint:errcheck

		return
	}

	server.Handle(ch, x11Display)
}
