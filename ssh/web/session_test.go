package web

import (
	"io"
	"testing"
	"testing/iotest"

	"github.com/shellhub-io/shellhub/ssh/web/mocks"
	"github.com/stretchr/testify/assert"
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
	mock := mocks.NewSocket(t)
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
	}, "expected redirToWs to panic when end is -1 and negative slice is attempted")
}

func TestRedirToWs_Regression_ZeroReadThenEOF(t *testing.T) {
	conn := &Conn{
		Socket: mocks.NewSocket(t),
	}

	reader := iotest.TimeoutReader(&zeroReadNoEOFReader{})

	assert.NotPanics(t, func() {
		_ = redirToWs(reader, conn)
	}, "expected redirToWs to handle zero read without panicking")
}
