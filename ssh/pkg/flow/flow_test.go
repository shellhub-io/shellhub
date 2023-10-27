package flow

import (
	"errors"
	"io"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/mocks"
	"github.com/stretchr/testify/assert"
)

func TestPipeIn(t *testing.T) {
	cases := []struct {
		description string
		client      io.Reader
		setup       func() io.Reader
		flow        *Flow
		expected    bool
	}{
		{
			description: "fails when there's an error on client read",
			client:      mocks.NewReader().OnRead(0, errors.New("read: error")),
			flow: &Flow{
				Stdin: mocks.NewWriteCloser().OnWrite(5, nil),
			},
			expected: false,
		},
		{
			description: "fails when there's an error on standard input write",
			client:      mocks.NewReader().OnRead(5, nil).OnRead(5, nil).EOF(),
			flow: &Flow{
				Stdin: mocks.NewWriteCloser().OnWrite(0, errors.New("write: error")),
			},
			expected: false,
		},
		{
			description: "succeeds when both client read and standard input write operations are successful",
			client:      mocks.NewReader().OnRead(5, nil).EOF(),
			flow: &Flow{
				Stdin: mocks.NewWriteCloser().OnWrite(5, nil),
			},
			expected: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			done := make(chan bool)

			go tc.flow.PipeIn(tc.client, done)
			assert.Equal(t, tc.expected, <-done)
		})
	}
}

func TestPipeOut(t *testing.T) {
	cases := []struct {
		description string
		client      io.Writer
		flow        *Flow
		expected    bool
	}{
		{
			description: "fails when there's an error on client write",
			client:      mocks.NewWriteCloser().OnWrite(0, errors.New("write: error")),
			flow: &Flow{
				Stdout: mocks.NewReader().OnRead(10, nil).EOF(),
			},
			expected: false,
		},
		{
			description: "fails when there's an error on standard output read",
			client:      mocks.NewWriteCloser().OnWrite(0, nil),
			flow: &Flow{
				Stdout: mocks.NewReader().OnRead(0, errors.New("read: error")).EOF(),
			},
			expected: false,
		},
		{
			description: "succeeds when both client write and standard output read operations are successful",
			client:      mocks.NewWriteCloser().OnWrite(5, nil),
			flow: &Flow{
				Stdout: mocks.NewReader().OnRead(5, nil).EOF(),
			},
			expected: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			done := make(chan bool)

			go tc.flow.PipeOut(tc.client, done)
			assert.Equal(t, tc.expected, <-done)
		})
	}
}

func TestPipeErr(t *testing.T) {
	cases := []struct {
		description string
		client      io.Writer
		flow        *Flow
		expected    bool
	}{
		{
			description: "fails when there's an error on client write",
			client:      mocks.NewWriteCloser().OnWrite(0, errors.New("write: error")),
			flow: &Flow{
				Stderr: mocks.NewReader().OnRead(10, nil).EOF(),
			},
			expected: false,
		},
		{
			description: "fails when there's an error on standard error read",
			client:      mocks.NewWriteCloser().OnWrite(0, nil),
			flow: &Flow{
				Stderr: mocks.NewReader().OnRead(0, errors.New("read: error")).EOF(),
			},
			expected: false,
		},
		{
			description: "succeeds when both client write and standard error read operations are successful",
			client:      mocks.NewWriteCloser().OnWrite(5, nil),
			flow: &Flow{
				Stderr: mocks.NewReader().OnRead(5, nil).EOF(),
			},
			expected: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			done := make(chan bool)

			go tc.flow.PipeErr(tc.client, done)
			assert.Equal(t, tc.expected, <-done)
		})
	}
}

func TestClose(t *testing.T) {
	cases := []struct {
		description string
		flow        *Flow
		expected    error
	}{
		{
			description: "fails when there's an error on standard input close",
			flow: &Flow{
				Stdin: mocks.NewWriteCloser().OnClose(errors.New("close: error")),
			},
			expected: errors.New("close: error"),
		},
		{
			description: "succeeds when standard input close operation is successful",
			flow: &Flow{
				Stdin: mocks.NewWriteCloser().OnClose(nil),
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.flow.Close())
		})
	}
}
