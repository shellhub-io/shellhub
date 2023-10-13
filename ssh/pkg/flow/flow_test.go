package flow

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewFlow(_ *testing.T) {
	// TODO: Consider mocking gossh.Session to test NewFlow.
}

// **NOTICE**
// In the following tests, we won't be creating instances of Flow using the NewFlow function.
// Instead, we'll simulate its behavior using mock structures and interfaces
// to specifically test the behavior of the methods without the dependencies of actual SSH sessions.

// writeCloser wraps bytes.Buffer to implement the io.WriteCloser interface.
type writeCloser struct {
	bytes.Buffer
}

// Close does nothing since bytes.Buffer doesn't require explicit close.
// It's implemented to satisfy the io.Closer interface.
func (wc *writeCloser) Close() error {
	return nil
}

// errorReader is a mock structure that implements io.Reader and always returns an error.
type errorReader struct{}

// Read always returns a forced error to simulate read failures.
func (r *errorReader) Read(_ []byte) (n int, err error) {
	return 0, errors.New("error")
}

// errorWriter is a mock structure that implements io.Writer and always returns an error.
type errorWriter struct{}

// Write always returns a forced error to simulate write failures.
func (w *errorWriter) Write(_ []byte) (n int, err error) {
	return 0, errors.New("error")
}

func TestPipeIn(t *testing.T) {
	cases := []struct {
		description string
		client      io.Reader
		flow        *Flow
		expected    bool
	}{
		{
			description: "fails when an error occurs in client.Read()",
			client:      &errorReader{},
			flow: &Flow{
				Stdin: &writeCloser{},
			},
			expected: false,
		},
		{
			description: "succeeds when client.Read operates as expected",
			client:      strings.NewReader("some test data"),
			flow: &Flow{
				Stdin: &writeCloser{},
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
			description: "fails when an error occurs in client.Write()",
			client:      &errorWriter{},
			flow: &Flow{
				Stdout: strings.NewReader("data"),
			},
			expected: false,
		},
		{
			description: "succeeds when client.Write operates as expected",
			client:      &bytes.Buffer{},
			flow: &Flow{
				Stdout: strings.NewReader("data"),
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
			description: "fails when an error occurs in client.Write()",
			client:      &errorWriter{},
			flow: &Flow{
				Stderr: strings.NewReader("data"),
			},
			expected: false,
		},
		{
			description: "succeeds when client.Write operates as expected",
			client:      &bytes.Buffer{},
			flow: &Flow{
				Stderr: strings.NewReader("data"),
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
