package web

import (
	"testing"
	"testing/iotest"

	"github.com/shellhub-io/shellhub/ssh/web/mocks"
	"github.com/stretchr/testify/assert"
)

type zeroReadNoEOFReader struct{}

func (r *zeroReadNoEOFReader) Read(p []byte) (int, error) {
	return 0, nil
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
