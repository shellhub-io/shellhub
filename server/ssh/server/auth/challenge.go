package auth

import (
	"errors"
	"strings"

	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/ssh/pkg/challenge"
	"github.com/shellhub-io/shellhub/server/ssh/session"
	log "github.com/sirupsen/logrus"
	gossh "golang.org/x/crypto/ssh"
)

type (
	offerFunc     func(gossh.ConnMetadata, gossh.PublicKey) (*gossh.Permissions, error)
	challengeFunc func(gossh.ConnMetadata, gossh.KeyboardInteractiveChallenge) (*gossh.Permissions, error)
)

const maxChallengeAnswers = 3

// ApprovalChallenge asks the person at the terminal to approve the login in the
// console, and finishes it once they answer.
//
// The answer is a release, never an authorization: it says the decision is worth
// reading, and the decision itself is the approval the person confirmed in the
// console. An empty answer, which is what a client with no way to prompt sends
// back, therefore denies rather than waits.
func ApprovalChallenge(ctx gliderssh.Context, offer offerFunc) challengeFunc {
	return func(_ gossh.ConnMetadata, ask gossh.KeyboardInteractiveChallenge) (*gossh.Permissions, error) {
		logger := log.WithFields(log.Fields{"uid": ctx.SessionID(), "sshid": ctx.User()})

		sess, ok := session.AuthenticableSessionOrDrop(ctx)
		if !ok {
			return nil, session.ErrAccessDenied
		}

		kind, ok := sess.Challenged()
		if !ok {
			logger.Info("a connection with no key to approve reached the challenge")

			return nil, deny(ask, sess.NoKeyReason())
		}

		for attempt := range maxChallengeAnswers {
			answers, err := ask(challenge.Approval, instruction(sess, kind), []string{question(sess)}, []bool{!sess.Web})
			if err != nil {
				return nil, err
			}

			approver, err := sess.Confirm(ctx, answer(answers))
			if err == nil {
				if err := sess.Resume(ctx, approver); err != nil {
					logger.WithError(err).Warn("failed to finish an approved login")

					return nil, deny(ask, reasonOf(err))
				}

				logger.Info("an approved login finished authenticating")

				return ctx.Permissions().Permissions, nil
			}

			if retryable(err) && attempt < maxChallengeAnswers-1 {
				continue
			}

			logger.WithError(err).Info("an approval was refused")

			if retryable(err) {
				return nil, deny(ask, "Too many attempts. Connect again to get a new request.", offer)
			}

			return nil, deny(ask, reasonOf(err), offer)
		}

		return nil, session.ErrAccessDenied
	}
}

func deny(ask gossh.KeyboardInteractiveChallenge, reason string, next ...offerFunc) error {
	_, _ = ask(challenge.Denied, reason, []string{""}, []bool{false})

	if len(next) > 0 && next[0] != nil {
		return &gossh.PartialSuccessError{
			Next: gossh.ServerAuthCallbacks{ //nolint:exhaustruct // publickey is the only method a refused client may still try
				PublicKeyCallback: next[0],
			},
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
	case errors.Is(err, session.ErrPromptDismissed):
		return "This prompt was dismissed, or this client cannot answer it. Connect again to get a new request."
	case errors.Is(err, session.ErrAccessDenied):
		return "An access policy does not allow this login."
	default:
		return "This login cannot continue. Try connecting again."
	}
}

func answer(answers []string) string {
	if len(answers) == 0 {
		return ""
	}

	return strings.TrimSpace(answers[0])
}

func question(sess *session.Session) string {
	if sess.Web {
		return sess.ApprovalCode
	}

	return "Confirmation code: "
}

func instruction(sess *session.Session, kind models.SSHApprovalKind) string {
	if sess.Web {
		return ""
	}

	if kind == models.SSHApprovalReauth {
		return sess.ReauthInstruction()
	}

	return sess.EnrollInstruction()
}
