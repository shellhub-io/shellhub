package auth

import (
	"net"

	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/server/ssh/session"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

// resolveKeyAuth builds the [session.Auth] for a presented key. It is a lookup:
// identity mode resolves the key to an account, legacy mode uses the key as
// presented, and neither writes anything.
func resolveKeyAuth(ctx gliderssh.Context, sess *session.Session, publicKey gliderssh.PublicKey) (session.Auth, error) {
	if !sess.IsIdentityMode() {
		return session.AuthPublicKey(publicKey), nil
	}

	// In identity mode the presented key IS the identity: resolve its fingerprint
	// to an account (connect straight through) or hold the login open for a
	// browser approval that turns the key into one. The ephemeral mint to the
	// agent is unchanged; the user's key is never forwarded.
	//
	// Native and web alike. The web terminal holds its own registered key, so it
	// is a first-class identity resolved here like any other public key; password
	// authentication is disabled entirely in identity mode.
	return sess.ResolveKeyAuth(ctx, publicKey)
}

// obtain returns the session for a connection that has got far enough to
// authenticate, dropping the connection otherwise.
func obtain(ctx gliderssh.Context, logger *log.Entry) (*session.Session, bool) {
	sess, state := session.ObtainSession(ctx)
	if state < session.StateEvaluated {
		logger.Trace("failed to get the session from context on public key handler")

		conn, ok := ctx.Value("conn").(net.Conn)
		if ok {
			conn.Close()
		}

		return nil, false
	}

	return sess, true
}

// PublicKeyOffer decides whether a public key is acceptable.
//
// The SSH protocol lets a client ask about a key without signing anything, and
// x/crypto runs this for that query, so it stays a read: no approval is parked,
// no session is registered, and the agent is not contacted. Everything with a
// cost waits for [PublicKeyVerified].
//
// An unenrolled key is acceptable in identity mode — enrolling it is the point
// of the flow — so a miss is not refused here.
func PublicKeyOffer(ctx gliderssh.Context, publicKey gliderssh.PublicKey) bool {
	logger := log.WithFields(
		log.Fields{
			"uid":   ctx.SessionID(),
			"sshid": ctx.User(),
			"key":   ssh.MarshalAuthorizedKey(publicKey),
		})

	logger.Trace("trying to use public key authentication")

	sess, ok := obtain(ctx, logger)
	if !ok {
		return false
	}

	auth, err := resolveKeyAuth(ctx, sess, publicKey)
	if err != nil {
		logger.WithError(err).Warn("failed to resolve the identity for the public key")

		return false
	}

	if err := auth.Offer(sess); err != nil {
		logger.WithError(err).Warn("refused the offered public key")

		return false
	}

	return true
}

// PublicKeyVerified authenticates once the client has proven it holds the key.
//
// x/crypto calls it only after the signature checks out and only when the offer
// was accepted, so this is where the browser approval, the session registration
// and the connection to the agent belong. Failing here is an ordinary
// authentication failure and counts against MaxAuthTries.
func PublicKeyVerified(ctx gliderssh.Context, publicKey gliderssh.PublicKey) bool {
	logger := log.WithFields(
		log.Fields{
			"uid":   ctx.SessionID(),
			"sshid": ctx.User(),
			"key":   ssh.MarshalAuthorizedKey(publicKey),
		})

	sess, ok := obtain(ctx, logger)
	if !ok {
		return false
	}

	// Resolved again rather than carried over from the offer: x/crypto caches one
	// key at a time, so a client that offers several keys can have the callback
	// re-run in any order, and a stale carry-over would authenticate the wrong
	// one. The lookup is a single read.
	auth, err := resolveKeyAuth(ctx, sess, publicKey)
	if err != nil {
		logger.WithError(err).Warn("failed to resolve the identity for the public key")

		return false
	}

	if err := sess.Auth(ctx, auth); err != nil {
		logger.WithError(err).Warn("failed to authenticate on device using public key")

		return false
	}

	logger.Info("succeeded to use public key authentication.")

	return true
}
