package web

import (
	"encoding/json"
	"io"
	"net"
	"testing"
	"testing/iotest"
	"time"

	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/server/ssh/pkg/banner"
	"github.com/shellhub-io/shellhub/server/ssh/pkg/challenge"
	"github.com/shellhub-io/shellhub/server/ssh/web/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gossh "golang.org/x/crypto/ssh"
)

type zeroReadNoEOFReader struct{}

func (r *zeroReadNoEOFReader) Read(p []byte) (int, error) {
	return 0, nil
}

type singleRead struct {
	data []byte
	read bool
}

func (r *singleRead) Read(p []byte) (int, error) {
	if r.read {
		return 0, io.EOF
	}

	n := copy(p, r.data)
	r.read = true

	return n, nil
}

func TestRedirToWs_Regression_EndNegative(t *testing.T) {
	mock := mocks.NewMockSocket(t)
	mock.On("Write", []byte{}).Return(0, nil).Once()

	conn := NewConn(mock)

	reader := &singleRead{data: []byte{0x80, 0x81, 0x82}}

	assert.NotPanics(t, func() {
		_ = redirToWs(reader, conn)
	}, "redirToWs must not panic when end becomes -1 (all UTF-8 continuation bytes)")
}

func TestRedirToWs_Regression_ZeroReadThenEOF(t *testing.T) {
	conn := &Conn{
		Socket: mocks.NewMockSocket(t),
	}

	reader := iotest.TimeoutReader(&zeroReadNoEOFReader{})

	assert.NotPanics(t, func() {
		_ = redirToWs(reader, conn)
	}, "expected redirToWs to handle zero read without panicking")
}

// TestBannerNewBannerErrorUnrecognizedSetsKindNone verifies that an unrecognized
// banner string yields KindNone, and that mapBannerError then returns ErrConnect —
// the production-critical path that prevents arbitrary banner text from leaking to
// web clients.
func TestBannerNewBannerErrorUnrecognizedSetsKindNone(t *testing.T) {
	e := NewBannerError("totally unknown banner text")

	require.Equal(t, banner.KindNone, e.Kind())

	got := mapBannerError(e)

	assert.ErrorIs(t, got, ErrConnect, "expected ErrConnect for unrecognized banner, got %v", got)
}

func dialWithBanner(t *testing.T, bannerText string) error {
	t.Helper()

	l, err := new(net.ListenConfig).Listen(t.Context(), "tcp", "127.0.0.1:0")
	require.NoError(t, err)

	srv := &gliderssh.Server{
		BannerHandler: func(_ gliderssh.Context) string {
			return bannerText
		},
		Handler: func(s gliderssh.Session) {},
	}

	go srv.Serve(l) //nolint:errcheck

	t.Cleanup(func() {
		srv.Close() //nolint:errcheck
		l.Close()   //nolint:errcheck
	})

	_, dialErr := gossh.Dial("tcp", l.Addr().String(), &gossh.ClientConfig{ //nolint:exhaustruct
		User:            "test",
		Auth:            []gossh.AuthMethod{gossh.Password("test")},
		HostKeyCallback: gossh.InsecureIgnoreHostKey(), //nolint:gosec
		BannerCallback: func(message string) error {
			if message != "" {
				return NewBannerError(message)
			}

			return nil
		},
	})

	return dialErr
}

// TestBannerKindMapsToSentinel verifies that each banner.Kind maps to the expected
// sentinel error via the production path: a real SSH dial with a BannerCallback
// returning NewBannerError, followed by errors.As extraction and mapBannerError.
// This mirrors exactly what newSession does at session.go:173-189.
func TestBannerKindMapsToSentinel(t *testing.T) {
	cases := []struct {
		description string
		kind        banner.Kind
		want        error
	}{
		{
			description: "KindConnectionFailed maps to ErrConnect",
			kind:        banner.KindConnectionFailed,
			want:        ErrConnect,
		},
		{
			description: "KindAccessDenied maps to ErrAccessDenied",
			kind:        banner.KindAccessDenied,
			want:        ErrAccessDenied,
		},
		{
			description: "KindInvalidSSHID maps to ErrInvalidSSHID",
			kind:        banner.KindInvalidSSHID,
			want:        ErrInvalidSSHID,
		},
		{
			description: "KindNone (empty banner) produces no BannerError and the dial succeeds",
			kind:        banner.KindNone,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			dialErr := dialWithBanner(t, banner.Message(tc.kind))

			var e *BannerError

			if tc.kind == banner.KindNone {
				require.NoError(t, dialErr, "expected dial to succeed when banner is empty")
				assert.NotErrorAs(t, dialErr, &e, "expected no BannerError for empty banner")

				return
			}

			require.ErrorAs(t, dialErr, &e, "expected errors.As to extract *BannerError from dial error, got: %v", dialErr)

			got := mapBannerError(e)

			assert.ErrorIs(t, got, tc.want, "expected %v, got %v", tc.want, got)
		})
	}
}

// TestApprovalChallengeCarriesTheCodeToTheBrowser pins the contract between the
// gateway and this bridge: the gateway puts the approval code in the challenge's
// question, and the bridge forwards exactly that to the browser and answers with
// whatever the browser sends back.
//
// It calls the bridge's challenge callback directly. The gateway half of this
// contract is pinned separately, by TestWebChallengeCarriesTheCodeInTheQuestion
// in the auth package; neither catches a field mismatch alone.
func TestApprovalChallengeCarriesTheCodeToTheBrowser(t *testing.T) {
	const code = "WXYZ2K7Q"

	browser, bridge := net.Pipe()

	t.Cleanup(func() {
		browser.Close() //nolint:errcheck
		bridge.Close()  //nolint:errcheck
	})

	conn := &Conn{Socket: bridge}

	answered := make(chan []string, 1)

	go func() {
		answers, err := approvalChallenge(conn, new(refusal))(challenge.Approval, "", []string{code}, []bool{true})
		if err != nil {
			close(answered)

			return
		}

		answered <- answers
	}()

	forwarded := struct {
		Kind messageKind `json:"kind"`
		Data string      `json:"data"`
	}{}

	require.NoError(t, json.NewDecoder(browser).Decode(&forwarded))

	assert.Equal(t, messageKindReauth, forwarded.Kind, "the browser is told an approval is waiting")
	assert.Equal(t, code, forwarded.Data, "the code has to survive the trip, or the modal opens on nothing")

	_, err := (&Conn{Socket: browser}).WriteMessage(&Message{Kind: messageKindReauthDone, Data: "CONF7788"})
	require.NoError(t, err)

	select {
	case answers, ok := <-answered:
		require.True(t, ok, "the challenge failed instead of answering")
		require.Len(t, answers, 1)
		assert.Equal(t, "CONF7788", answers[0], "what the person confirmed is what answers the challenge")
	case <-time.After(5 * time.Second):
		require.Fail(t, "the bridge never answered the challenge")
	}
}

// TestApprovalChallengeRefusesAnEmptyCode covers the shape of the bug this test
// file exists to catch: a gateway that names the challenge correctly but leaves
// the code out would otherwise open a modal on an empty string.
func TestApprovalChallengeRefusesAnEmptyCode(t *testing.T) {
	_, bridge := net.Pipe()

	t.Cleanup(func() { bridge.Close() }) //nolint:errcheck

	_, err := approvalChallenge(&Conn{Socket: bridge}, new(refusal))(challenge.Approval, "", []string{""}, []bool{true})

	require.ErrorIs(t, err, ErrApprovalCodeMissing)
}

// TestDeniedChallengeIsNotAnApproval verifies the bridge does not treat a
// refusal as something the person can answer: a modal opened here would wait on
// a decision the gateway has already made.
func TestDeniedChallengeIsNotAnApproval(t *testing.T) {
	_, bridge := net.Pipe()

	t.Cleanup(func() { bridge.Close() }) //nolint:errcheck

	answers, err := approvalChallenge(&Conn{Socket: bridge}, new(refusal))(challenge.Denied, "a reason", []string{""}, []bool{false})

	require.NoError(t, err)
	assert.Equal(t, []string{""}, answers, "a refusal is answered with nothing, not forwarded to the browser")
}

// TestDeniedChallengeCarriesTheReasonToTheBrowser is why the refusal is recorded
// rather than only logged. A refusal has no answer the gateway reads, so the
// dial that follows fails as a plain authentication error, and the browser was
// being told its key is not authorized for the device on a login the person had
// just rejected themselves.
func TestDeniedChallengeCarriesTheReasonToTheBrowser(t *testing.T) {
	_, bridge := net.Pipe()

	t.Cleanup(func() { bridge.Close() }) //nolint:errcheck

	refused := new(refusal)

	require.NoError(t, refused.Err(), "a login nobody refused carries no reason")

	_, err := approvalChallenge(&Conn{Socket: bridge}, refused)(
		challenge.Denied, "This login was rejected in the console.", []string{""}, []bool{false})
	require.NoError(t, err)

	got := refused.Err()

	require.ErrorIs(t, got, ErrApprovalRefused)
	assert.Equal(t, "the login approval was refused: This login was rejected in the console.", got.Error(),
		"the console matches this exact prefix to render the reason, so the whole string is the contract")
}
