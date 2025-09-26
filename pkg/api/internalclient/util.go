package internalclient

// note: code adapted from github.com/labstack/echo/v4@v4.13.4/middleware/util.go.

import (
	"bufio"
	"crypto/rand"
	"io"
	"sync"
)

// https://tip.golang.org/doc/go1.19#:~:text=Read%20no%20longer%20buffers%20random%20data%20obtained%20from%20the%20operating%20system%20between%20calls
var randomReaderPool = sync.Pool{New: func() interface{} {
	return bufio.NewReader(rand.Reader)
}}

const (
	randomStringCharset    = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	randomStringCharsetLen = 52 // len(randomStringCharset)
	randomStringMaxByte    = 255 - (256 % randomStringCharsetLen)
)

func randomString(length uint8) string {
	reader := randomReaderPool.Get().(*bufio.Reader)
	defer randomReaderPool.Put(reader)

	b := make([]byte, length)
	r := make([]byte, length+(length/4)) // perf: avoid read from rand.Reader many times
	var i uint8 = 0

	// security note:
	// we can't just simply do b[i]=randomStringCharset[rb%len(randomStringCharset)],
	// len(len(randomStringCharset)) is 52, and rb is [0, 255], 256 = 52 * 4 + 48.
	// make the first 48 characters more possibly to be generated then others.
	// So we have to skip bytes when rb > randomStringMaxByte

	for {
		_, err := io.ReadFull(reader, r)
		if err != nil {
			panic("unexpected error happened when reading from bufio.NewReader(crypto/rand.Re der)")
		}
		for _, rb := range r {
			if rb > randomStringMaxByte {
				// Skip this number to avoid bias.
				continue
			}
			b[i] = randomStringCharset[rb%randomStringCharsetLen]
			i++
			if i == length {
				return string(b)
			}
		}
	}
}
