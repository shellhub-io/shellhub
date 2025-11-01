package channels

import (
	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/pkg/srdp"
	"github.com/shellhub-io/shellhub/pkg/srdp/displays"
	log "github.com/sirupsen/logrus"
	gossh "golang.org/x/crypto/ssh"
)

type SRDPChannel struct {
	srdp *srdp.SRDPServer
}

func NewSRDPChannel(logger *log.Entry) (*SRDPChannel, error) {
	// TODO: Add authentication methods as needed.
	server := srdp.NewSRDPServer(&srdp.Config{
		Name:   "SRDP Server",
		Auths:  []srdp.Auth{srdp.NewNoAuth()},
		Logger: logger,
	})

	return &SRDPChannel{
		srdp: server,
	}, nil
}

func (c *SRDPChannel) SRDPHandler(srv *gliderssh.Server, conn *gossh.ServerConn, newChan gossh.NewChannel, ctx gliderssh.Context) {
	logger := log.WithFields(log.Fields{
		"user":    ctx.User(),
		"channel": newChan.ChannelType(),
		"handler": "SRDPHandler",
	})

	logger.Info("New SRDP channel request received")
	defer logger.Info("SRDP channel handler finished processing")

	d := newChan.ExtraData()

	var data Data
	if err := gossh.Unmarshal(d, &data); err != nil {
		logger.WithError(err).Error("Failed to unmarshal SRDP channel data")

		newChan.Reject(gossh.Prohibited, "invalid channel data") //nolint:errcheck

		return
	}

	ch, reqs, err := newChan.Accept()
	if err != nil {
		logger.WithError(err).Error("Failed to accept SRDP channel")

		newChan.Reject(gossh.ConnectionFailed, "could not accept channel") //nolint:errcheck

		return
	}

	// NOTE: Discard all global out-of-band Requests.
	go gossh.DiscardRequests(reqs)

	display, err := displays.NewX11Display(data.Display)
	if err != nil {
		logger.WithError(err).Error("Failed to create X11 display")

		newChan.Reject(gossh.ConnectionFailed, "failed to create X11 display") //nolint:errcheck

		return
	}
	defer display.Close()

	if err := c.srdp.Handle(ch, display); err != nil {
		logger.WithError(err).Error("Failed to handle SRDP connection")

		newChan.Reject(gossh.ConnectionFailed, "failed to handle SRDP connection") //nolint:errcheck

		return
	}
}
