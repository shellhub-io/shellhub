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

func DefaultVNCHandler() gliderssh.ChannelHandler {
	return func(srv *gliderssh.Server, conn *gossh.ServerConn, newChan gossh.NewChannel, ctx gliderssh.Context) {
		log.WithFields(log.Fields{
			"username": ctx.User(),
		}).Info("handling vnc channel")
		defer log.WithFields(log.Fields{
			"username": ctx.User(),
		}).Info("vnc channel has done")

		sess, _ := session.ObtainSession(ctx)
		go func() {
			// NOTICE: As [gossh.ServerConn] is shared by all channels calls, close it after a channel close block any
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

		logger.Info("new vnc channel request")

		d := newChan.ExtraData()

		// data := new(Data)
		// if err := gossh.Unmarshal(d, data); err != nil && len(d) > 0 {
		// 	reject(ctx, err, "failed to parse vnc channel data", newChan)

		// 	return
		// }

		// TODO: Validate data fields.

		client, err := sess.NewClientChannel(newChan, seat)
		if err != nil {
			reject(ctx, err, "failed to accept the channel opening", newChan)

			return
		}

		defer client.Close()

		logger.Info("new vnc client request")

		agent, err := sess.NewAgentChannelWithData(VNCChannel, seat, d)
		if err != nil {
			reject(ctx, err, "failed to open the session channel on agent", newChan)

			return
		}

		defer agent.Close()

		logger.Info("vnc channel accepted")

		wg := new(sync.WaitGroup)

		wg.Add(1)
		go func() {
			defer wg.Done()
			defer logger.Trace("vnc channel data copy done")

			if _, err := io.Copy(client.Channel, agent.Channel); err != nil && err != io.EOF {
				log.WithError(err).Error("failed to pipe agent connection")
			}
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			defer logger.Trace("vnc channel data copy done")

			if _, err := io.Copy(agent.Channel, client.Channel); err != nil && err != io.EOF {
				log.WithError(err).Error("failed to pipe agent connection")
			}
		}()

		logger.Trace("vnc channel data piping started")

		wg.Wait()

		logger.Info("vnc session has ended")
	}
}
