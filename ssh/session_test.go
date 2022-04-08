package main

import (
	"testing"

	"github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/ssh/mocks"
	"github.com/stretchr/testify/assert"
)

func TestHandlePty(t *testing.T) {
	t.Run("HandleWS", func(t *testing.T) {
		sessionMock := &mocks.Session{}
		session := &Session{session: sessionMock}

		sessionMock.On("Environ").Return([]string{"WS=true"}).Once()
		sessionMock.On("Pty").Return(ssh.Pty{Term: "xterm.js", Window: ssh.Window{Width: 100, Height: 40}}, nil, true).Once()
		handlePty(session)
		assert.Equal(t, true, session.Pty)
		assert.Equal(t, Web, session.Type)
		assert.Equal(t, "xterm.js", session.Term)

		sessionMock.AssertExpectations(t)
	})

	t.Run("HandleIterativePty", func(t *testing.T) {
		sessionMock := &mocks.Session{}
		session := &Session{session: sessionMock}

		sessionMock.On("Environ").Return([]string{"WS=false"}).Once()
		sessionMock.On("Pty").Return(ssh.Pty{Term: "xterm", Window: ssh.Window{Width: 100, Height: 40}}, nil, true).Once()
		sessionMock.On("Command").Return([]string{}).Once()
		handlePty(session)
		assert.Equal(t, true, session.Pty)
		assert.Equal(t, Term, session.Type)
		assert.Equal(t, "xterm", session.Term)

		sessionMock.AssertExpectations(t)
	})

	t.Run("HandleNotIterativePty", func(t *testing.T) {
		sessionMock := &mocks.Session{}
		session := &Session{session: sessionMock}

		sessionMock.On("Environ").Return([]string{"WS=false"}).Once()
		sessionMock.On("Pty").Return(ssh.Pty{}, nil, false).Once()
		sessionMock.On("Command").Return([]string{"ls"}).Once()
		handlePty(session)
		assert.Equal(t, false, session.Pty)
		assert.Equal(t, Exec, session.Type)
		assert.Equal(t, "", session.Term)

		sessionMock.AssertExpectations(t)
	})

	t.Run("HandleSCP", func(t *testing.T) {
		sessionMock := &mocks.Session{}
		session := &Session{session: sessionMock}

		sessionMock.On("Environ").Return([]string{"WS=false"}).Once()
		sessionMock.On("Pty").Return(ssh.Pty{}, nil, false).Once()
		sessionMock.On("Command").Return([]string{"scp"}).Once()
		handlePty(session)
		assert.Equal(t, false, session.Pty)
		assert.Equal(t, SCP, session.Type)
		assert.Equal(t, "", session.Term)

		sessionMock.AssertExpectations(t)
	})
}
