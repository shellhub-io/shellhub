package auth

import (
	"errors"
	"strings"

	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/ssh/session"
	log "github.com/sirupsen/logrus"
	gossh "golang.org/x/crypto/ssh"
)

// offerFunc and challengeFunc are the shapes x/crypto's ServerAuthCallbacks
// holds. They are declared here because the package exports the fields, not the
// types.
type (
	offerFunc     func(gossh.ConnMetadata, gossh.PublicKey) (*gossh.Permissions, error)
	challengeFunc func(gossh.ConnMetadata, gossh.KeyboardInteractiveChallenge) (*gossh.Permissions, error)
)

// maxChallengeAnswers bounds how many times one prompt is re-asked within a
// single challenge. OpenSSH starts keyboard-interactive at most
// NumberOfPasswordPrompts times per connection (three by default, and a user may
// set one), so re-asking happens here rather than by failing and hoping the
// client comes back.
const maxChallengeAnswers = 3

// challengeName is what the client shows above the prompt. The web terminal
// bridge matches on it exactly, so it is a stable identifier and not prose.
const challengeName = "shellhub-approval"

// ApprovalChallenge asks the person at the terminal to approve the login in the
// console, and finishes it once they answer.
//
// The answer is a release, never an authorization: it says the decision is worth
// reading, and the decision itself is the approval the person confirmed in the
// console. An empty answer, which is what a client with no way to prompt sends
// back, therefore denies rather than waits.
func ApprovalChallenge(ctx gliderssh.Context, offer offerFunc) challengeFunc {
	return func(_ gossh.ConnMetadata, challenge gossh.KeyboardInteractiveChallenge) (*gossh.Permissions, error) {
		logger := log.WithFields(log.Fields{"uid": ctx.SessionID(), "sshid": ctx.User()})

		sess, ok := session.AuthenticableSessionOrDrop(ctx)
		if !ok {
			return nil, session.ErrAccessDenied
		}

		kind, ok := sess.Challenged()
		if !ok {
			return nil, session.ErrAccessDenied
		}

		for attempt := range maxChallengeAnswers {
			answers, err := challenge(challengeName, instruction(sess, kind), []string{prompt(sess)}, []bool{!sess.Web})
			if err != nil {
				return nil, err
			}

			approver, err := sess.Confirm(ctx, answer(answers))
			if err == nil {
				if err := sess.Resume(ctx, approver); err != nil {
					logger.WithError(err).Warn("failed to finish an approved login")

					return nil, deny(challenge, reasonOf(err))
				}

				logger.Info("an approved login finished authenticating")

				return ctx.Permissions().Permissions, nil
			}

			if retryable(err) && attempt < maxChallengeAnswers-1 {
				continue
			}

			logger.WithError(err).Info("an approval was refused")

			if retryable(err) {
				return nil, deny(challenge, "Too many attempts. Connect again to get a new request.", offer)
			}

			return nil, deny(challenge, reasonOf(err), offer)
		}

		return nil, session.ErrAccessDenied
	}
}

// deny tells the person why, then refuses. The reason rides on one last prompt
// because that is the only channel that reaches a client running under `ssh -q`,
// which prints nothing else the server sends.
//
// When next is given, the refusal hands the remaining authentication methods
// back to the client instead of ending the conversation.
func deny(challenge gossh.KeyboardInteractiveChallenge, reason string, next ...offerFunc) error {
	_, _ = challenge(challengeName, "", []string{reason + "\r\nPress Enter to disconnect. "}, []bool{true})

	if len(next) > 0 && next[0] != nil {
		return &gossh.PartialSuccessError{
			Next: gossh.ServerAuthCallbacks{PublicKeyCallback: next[0]}, //nolint:exhaustruct
		}
	}

	return session.ErrAccessDenied
}

func retryable(err error) bool {
	return errors.Is(err, session.ErrApprovalPending) || errors.Is(err, session.ErrConfirmationMismatch)
}

func reasonOf(err error) string {
	switch {
	case errors.Is(err, session.ErrApprovalRejected):
		return "This login was rejected in the console."
	case errors.Is(err, session.ErrApprovalExpired):
		return "This request expired. Connect again to get a new one."
	case errors.Is(err, session.ErrApprovalPending):
		return "This login was never approved in the console."
	case errors.Is(err, session.ErrConfirmationMismatch):
		return "That confirmation code is not the one the console showed."
	case errors.Is(err, session.ErrAccessDenied):
		return "An access policy does not allow this login."
	default:
		return "This login cannot continue."
	}
}

func answer(answers []string) string {
	if len(answers) == 0 {
		return ""
	}

	return strings.TrimSpace(answers[0])
}

func prompt(sess *session.Session) string {
	if sess.Web {
		return ""
	}

	return "Confirmation code: "
}

// instruction is what the client renders above the prompt. A web session gets
// the bare code, because the bridge on the other side is a program; a terminal
// gets the text a person reads.
func instruction(sess *session.Session, kind models.SSHApprovalKind) string {
	if sess.Web {
		return sess.ApprovalCode
	}

	if kind == models.SSHApprovalReauth {
		return sess.ReauthInstruction()
	}

	return sess.EnrollInstruction()
}
