package web

import (
	"errors"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
)

type recordingSession struct {
	calls []string
	env   map[string]string
	term  string
	rows  int
	cols  int

	setenvErr     error
	requestPtyErr error
	shellErr      error
}

func newRecordingSession() *recordingSession {
	return &recordingSession{env: make(map[string]string)}
}

func (r *recordingSession) Setenv(name, value string) error {
	r.calls = append(r.calls, "setenv")
	r.env[name] = value

	return r.setenvErr
}

func (r *recordingSession) RequestPty(term string, height, width int, _ ssh.TerminalModes) error {
	r.calls = append(r.calls, "pty")
	r.term = term
	r.rows = height
	r.cols = width

	return r.requestPtyErr
}

func (r *recordingSession) Shell() error {
	r.calls = append(r.calls, "shell")

	return r.shellErr
}

func TestPrepareShell(t *testing.T) {
	dim := Dimensions{Rows: 24, Cols: 80}

	cases := []struct {
		description   string
		session       func() *recordingSession
		expectedCalls []string
		expectedErr   error
	}{
		{
			description:   "announces the character locale before starting the shell",
			session:       newRecordingSession,
			expectedCalls: []string{"setenv", "pty", "shell"},
			expectedErr:   nil,
		},
		{
			description: "continues the bring-up when the env request is rejected",
			session: func() *recordingSession {
				session := newRecordingSession()
				session.setenvErr = errors.New("env request rejected")

				return session
			},
			expectedCalls: []string{"setenv", "pty", "shell"},
			expectedErr:   nil,
		},
		{
			description: "fails when the pty request fails",
			session: func() *recordingSession {
				session := newRecordingSession()
				session.requestPtyErr = errors.New("pty request failed")

				return session
			},
			expectedCalls: []string{"setenv", "pty"},
			expectedErr:   ErrPty,
		},
		{
			description: "fails when the shell request fails",
			session: func() *recordingSession {
				session := newRecordingSession()
				session.shellErr = errors.New("shell request failed")

				return session
			},
			expectedCalls: []string{"setenv", "pty", "shell"},
			expectedErr:   ErrShell,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			session := tc.session()

			err := prepareShell(log.WithField("test", tc.description), session, dim)

			require.ErrorIs(t, err, tc.expectedErr)
			assert.Equal(t, tc.expectedCalls, session.calls)
			assert.Equal(t, map[string]string{"LC_CTYPE": "C.UTF-8"}, session.env)
		})
	}
}

func TestPrepareShellRequestsThePtyWithTheClientDimensions(t *testing.T) {
	session := newRecordingSession()

	require.NoError(t, prepareShell(log.WithField("test", t.Name()), session, Dimensions{Rows: 40, Cols: 132}))

	assert.Equal(t, "xterm", session.term)
	assert.Equal(t, 40, session.rows)
	assert.Equal(t, 132, session.cols)
}
