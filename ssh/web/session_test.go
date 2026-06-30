package web

import (
	"errors"
	"io"
	"net"
	"testing"
	"testing/iotest"

	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/ssh/pkg/banner"
	"github.com/shellhub-io/shellhub/ssh/web/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gossh "golang.org/x/crypto/ssh"
)

type zeroReadNoEOFReader struct{}

func (r *zeroReadNoEOFReader) Read(p []byte) (int, error) {
	return 0, nil
}

// singleRead returns the provided bytes on the first Read call, then EOF.
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

	// All three bytes are UTF-8 continuation bytes, which will cause the
	// logic in redirToWs to set end to -1 if not handled properly.
	// This test ensures that the function does not panic in such a case.
	//
	// https://datatracker.ietf.org/doc/html/rfc3629#section-3
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

// TestBannerNewBannerErrorSetsKind verifies that NewBannerError populates the Kind
// by calling banner.Classify on the message at construction time.
func TestBannerNewBannerErrorSetsKind(t *testing.T) {
	cases := []struct {
		description string
		kind        banner.Kind
	}{
		{
			description: "KindConnectionFailed",
			kind:        banner.KindConnectionFailed,
		},
		{
			description: "KindAccessDenied",
			kind:        banner.KindAccessDenied,
		},
		{
			description: "KindInvalidSSHID",
			kind:        banner.KindInvalidSSHID,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			msg := banner.Message(tc.kind)
			e := NewBannerError(msg)

			assert.Equal(t, tc.kind, e.Kind())
			assert.Equal(t, msg, e.Message, "NewBannerError must preserve the raw message verbatim")
			assert.Equal(t, msg, e.Error(), "Error() must return the raw message verbatim")
		})
	}
}

// TestBannerNewBannerErrorUnrecognizedSetsKindNone verifies that an unrecognized
// banner string yields KindNone, and that mapBannerError then returns ErrConnect —
// the production-critical path that prevents arbitrary banner text from leaking to
// web clients.
func TestBannerNewBannerErrorUnrecognizedSetsKindNone(t *testing.T) {
	e := NewBannerError("totally unknown banner text")

	require.Equal(t, banner.KindNone, e.Kind())

	got := mapBannerError(e)

	assert.True(t, errors.Is(got, ErrConnect), "expected ErrConnect for unrecognized banner, got %v", got)
}

// TestMapBannerErrorEmptyMessage verifies that mapBannerError returns ErrConnect
// for a BannerError with an empty Message (KindNone with no text). This is the
// default branch of mapBannerError when no log entry is emitted, which is
// distinct from the non-empty-unknown-text path in
// TestBannerNewBannerErrorUnrecognizedSetsKindNone.
func TestMapBannerErrorEmptyMessage(t *testing.T) {
	e := NewBannerError("")

	require.Equal(t, banner.KindNone, e.Kind())

	got := mapBannerError(e)

	assert.True(t, errors.Is(got, ErrConnect), "expected ErrConnect for empty-message BannerError, got %v", got)
}

// dialWithBanner starts a gliderssh server that sends the given banner, then
// dials it with a BannerCallback that returns NewBannerError — exactly as the
// production newSession does. It returns the dial error (which embeds a
// *BannerError) so callers can exercise the errors.As path.
func dialWithBanner(t *testing.T, bannerText string) error {
	t.Helper()

	l, err := net.Listen("tcp", "127.0.0.1:0")
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
			// KindNone produces an empty banner string. The BannerCallback returns nil
			// for an empty string, so no *BannerError is injected into the dial error.
			// The server has no auth handlers set, so it accepts any credentials and
			// the dial succeeds: there is no dial error at all. mapBannerError is never
			// called for this case, so there is no sentinel to assert — `want` is
			// intentionally absent. The unknown-banner → ErrConnect sentinel mapping is
			// covered separately by TestBannerNewBannerErrorUnrecognizedSetsKindNone.
			description: "KindNone (empty banner) produces no BannerError and the dial succeeds",
			kind:        banner.KindNone,
			// want is deliberately omitted: mapBannerError is not reached on this path.
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			dialErr := dialWithBanner(t, banner.Message(tc.kind))

			// The production newSession path: errors.As extracts *BannerError from
			// the dial error, then mapBannerError converts it to a sentinel.
			var e *BannerError

			if tc.kind == banner.KindNone {
				// Empty banner: BannerCallback returns nil, so the dial should
				// succeed with no error and no *BannerError embedded.
				require.NoError(t, dialErr, "expected dial to succeed when banner is empty")
				assert.False(t, errors.As(dialErr, &e), "expected no BannerError for empty banner")

				return
			}

			require.True(t, errors.As(dialErr, &e), "expected errors.As to extract *BannerError from dial error, got: %v", dialErr)

			got := mapBannerError(e)

			assert.True(t, errors.Is(got, tc.want), "expected %v, got %v", tc.want, got)
		})
	}
}

// TestBannerWireClassify verifies the full production round-trip: an SSH server
// whose BannerHandler returns banner.Message(k) delivers that string over the wire;
// the client's BannerCallback (identical to newSession's) returns NewBannerError;
// gossh.Dial surfaces it as a dial error; and errors.As correctly extracts a
// *BannerError with the expected Kind. This is the behavior that newSession depends on.
func TestBannerWireClassify(t *testing.T) {
	cases := []struct {
		description string
		kind        banner.Kind
	}{
		{
			description: "KindConnectionFailed round-trip over wire",
			kind:        banner.KindConnectionFailed,
		},
		{
			description: "KindAccessDenied round-trip over wire",
			kind:        banner.KindAccessDenied,
		},
		{
			description: "KindInvalidSSHID round-trip over wire",
			kind:        banner.KindInvalidSSHID,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			dialErr := dialWithBanner(t, banner.Message(tc.kind))

			require.Error(t, dialErr, "expected dial to fail when BannerCallback returns a non-nil error")

			var e *BannerError

			require.True(t, errors.As(dialErr, &e), "expected errors.As to extract *BannerError from dial error, got: %v", dialErr)

			assert.Equal(t, tc.kind, e.Kind(), "Kind extracted over the wire must match the banner sent by the server")
		})
	}
}
