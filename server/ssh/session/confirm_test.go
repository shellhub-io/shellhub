package session

import (
	"context"
	"errors"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/services"
	servicemocks "github.com/shellhub-io/shellhub/server/api/services/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func parkedSession(service *servicemocks.MockService) *Session {
	sess := newTestSession(service, nil)
	sess.ApprovalCode = "WXYZ2K7Q"
	sess.challenge = &challenge{auth: nil, kind: models.SSHApprovalIdentity}

	return sess
}

// TestConfirmTreatsNoAnswerOnAnUndecidedLoginAsDismissed is what keeps a dismissed dialog from
// reopening. An empty answer used to be read as "not approved yet", which the
// prompt retries, so closing the web terminal's dialog brought it straight back
// up to the retry limit. There is nothing to look up for an empty answer, and
// nothing a second prompt could get that the first did not.
func TestConfirmTreatsNoAnswerOnAnUndecidedLoginAsDismissed(t *testing.T) {
	for _, answer := range []string{"", "   ", "\n"} {
		t.Run("answer "+answer, func(t *testing.T) {
			service := servicemocks.NewMockService(t)
			service.On("GetSSHApprovalStatus", mock.Anything, mock.Anything).
				Return(&models.SSHApprovalStatus{State: models.SSHApprovalPending}, nil) //nolint:exhaustruct // only the state is read

			_, err := parkedSession(service).Confirm(context.Background(), answer)

			require.ErrorIs(t, err, ErrPromptDismissed)
			require.NotErrorIs(t, err, ErrApprovalPending, "a dismissal must not read as something worth asking again for")
		})
	}
}

// TestConfirmReportsAnExplicitRejection is why the dismissal check reads the
// decision first. Rejecting in the console answers the challenge with the same
// empty frame dismissing the dialog does, so the two are indistinguishable on
// the wire; the stored decision is what tells them apart, and someone who has
// just clicked Reject should be told that rather than that they dismissed a
// prompt.
func TestConfirmReportsAnExplicitRejection(t *testing.T) {
	service := servicemocks.NewMockService(t)
	service.On("GetSSHApprovalStatus", mock.Anything, mock.Anything).
		Return(&models.SSHApprovalStatus{State: models.SSHApprovalRejected}, nil) //nolint:exhaustruct // only the state is read

	_, err := parkedSession(service).Confirm(context.Background(), "")

	require.ErrorIs(t, err, ErrApprovalRejected)
	require.NotErrorIs(t, err, ErrPromptDismissed, "a rejection the person made is not a dismissal")
}

// TestConfirmStillChecksAWrongCode pins the other side of that line: an answer
// that is not empty is a code to check, and a wrong one is a mismatch the
// prompt may retry, not a dismissal.
func TestConfirmStillChecksAWrongCode(t *testing.T) {
	service := servicemocks.NewMockService(t)
	service.On("GetSSHApprovalStatus", mock.Anything, mock.Anything).
		Return(&models.SSHApprovalStatus{State: models.SSHApprovalConfirmed, UserID: "user-id", ConfirmationCode: "REALCODE"}, nil) //nolint:exhaustruct // only the fields Confirm reads

	_, err := parkedSession(service).Confirm(context.Background(), "WRONGONE")

	require.ErrorIs(t, err, ErrConfirmationMismatch)
	require.NotErrorIs(t, err, ErrPromptDismissed)
}

// TestConfirmReadsAVanishedApprovalAsExpired pins the message a person gets for
// taking too long: the approval row is gone once its TTL passes, and the service
// reports that as not found. Without the mapping that fell through to the generic
// refusal, telling someone who only needed to reconnect that the login was broken.
func TestConfirmReadsAVanishedApprovalAsExpired(t *testing.T) {
	service := servicemocks.NewMockService(t)
	service.On("GetSSHApprovalStatus", mock.Anything, mock.Anything).
		Return(nil, services.NewErrSSHApprovalCodeNotFound("WXYZ2K7Q", errors.New("no rows")))

	_, err := parkedSession(service).Confirm(context.Background(), "ANYCODE1")

	require.ErrorIs(t, err, ErrApprovalExpired)
	require.NotErrorIs(t, err, ErrApprovalPending, "an expired approval is not one still worth waiting on")
}
