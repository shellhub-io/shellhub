package web

import (
	"bytes"
	"testing"
	"testing/iotest"

	"github.com/stretchr/testify/assert"
)

type zeroReadNoEOFReader struct{}

func (r *zeroReadNoEOFReader) Read(p []byte) (int, error) {
	return 0, nil
}

func TestRedirToWs_Regression_ZeroReadThenEOF(t *testing.T) {
	reader := iotest.TimeoutReader(&zeroReadNoEOFReader{})
	writer := iotest.NewWriteLogger("test", &bytes.Buffer{})

	assert.NotPanics(t, func() {
		_ = redirToWs(reader, writer)
	}, "expected redirToWs to handle zero read without panicking")
}
