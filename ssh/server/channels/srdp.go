package channels

import (
	"io"
	"sync"

	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/ssh/session"
	log "github.com/sirupsen/logrus"
	gossh "golang.org/x/crypto/ssh"
)

func reject(ctx gliderssh.Context, err error, msg string, newChan gossh.NewChannel) {
	log.WithError(err).WithFields(log.Fields{
		"username": ctx.User(),
		"channel":  newChan.ChannelType(),
	}).Error(msg)

	newChan.Reject(gossh.ConnectionFailed, msg) //nolint:errcheck
}

func DefaultSRDPHandler() gliderssh.ChannelHandler {
	return func(srv *gliderssh.Server, conn *gossh.ServerConn, newChan gossh.NewChannel, ctx gliderssh.Context) {
		log.WithFields(log.Fields{
			"username": ctx.User(),
		}).Info("handling srdp channel")
		defer log.WithFields(log.Fields{
			"username": ctx.User(),
		}).Info("srdp channel has done")

		sess, _ := session.ObtainSession(ctx)
		go func() {
			// NOTE: As [gossh.ServerConn] is shared by all channels calls, close it after a channel close block any
			// other channel involkation. To avoid it, we wait for the connection be closed to finish the sesison.
			conn.Wait() //nolint:errcheck

			sess.Finish() //nolint:errcheck
		}()

		logger := log.WithFields(
			log.Fields{
				"uid":      sess.UID,
				"sshid":    sess.SSHID,
				"device":   sess.Device.UID,
				"username": sess.Target.Username,
				"ip":       sess.IPAddress,
			})

		seat, err := sess.NewSeat()
		if err != nil {
			reject(ctx, err, "failed to create a new seat on the SSH session", newChan)

			return
		}

		logger = logger.WithField("seat", seat)

		logger.Trace("accepting srdp channel")

		client, err := sess.NewClientChannel(newChan, seat)
		if err != nil {
			reject(ctx, err, "failed to accept the channel opening", newChan)

			return
		}

		defer client.Close()

		logger.Trace("srdp channel accepted")

		d := newChan.ExtraData()

		logger.Trace("opening srdp channel on agent")

		agent, err := sess.NewAgentChannelWithData(SRDPChannel, seat, d)
		if err != nil {
			reject(ctx, err, "failed to open the session channel on agent", newChan)

			return
		}

		defer agent.Close()

		logger.Trace("srdp channel opened on agent")

		wg := new(sync.WaitGroup)

		wg.Add(1)
		go func() {
			defer wg.Done()
			defer logger.Trace("srdp channel data copy done")

			logger.Trace("starting srdp channel data copy from agent to client")

			if _, err := io.Copy(client.Channel, agent.Channel); err != nil && err != io.EOF {
				log.WithError(err).Error("failed to copy data to client channel")
			}
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			defer logger.Trace("srdp channel data copy done")

			logger.Trace("starting srdp channel data copy from client to agent")

			if _, err := io.Copy(agent.Channel, client.Channel); err != nil && err != io.EOF {
				log.WithError(err).Error("failed to copy data to agent channel")
			}
		}()

		wg.Wait()
	}
}
