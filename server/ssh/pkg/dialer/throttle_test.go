package dialer

import (
	"bytes"
	"io"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func expectedMinDuration(total, bps, burst int) time.Duration {
	if bps <= 0 {
		return 0
	}

	remaining := total - burst
	if remaining <= 0 {
		return 0
	}

	secs := float64(remaining) / float64(bps)

	return time.Duration(secs * float64(time.Second))
}

func TestThrottler_TableDriven(t *testing.T) {
	cases := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "UnlimitedReadFast",
			run: func(t *testing.T) {
				data := bytes.Repeat([]byte("x"), 1024)
				r := bytes.NewReader(data)

				th := NewThrottler(r, nil) // no limits

				buf := make([]byte, len(data))
				start := time.Now()
				n, err := th.Read(buf)
				dur := time.Since(start)

				assert.Truef(t, err == nil || err == io.EOF, "unexpected read error: %v", err)
				assert.Equal(t, len(data), n, "read bytes mismatch")
				assert.LessOrEqual(t, dur, 100*time.Millisecond, "unlimited read took too long")
			},
		},
		{
			name: "NegativeLimitValidation",
			run: func(t *testing.T) {
				th := NewThrottler(nil, nil)
				err := th.UpdateReadLimit(-1, 1)
				assert.Equal(t, ErrNegativeLimit, err)
				err = th.UpdateWriteLimit(-1, 1)
				assert.Equal(t, ErrNegativeLimit, err)
			},
		},
		{
			name: "ReadRateEnforced",
			run: func(t *testing.T) {
				total := 200
				bps := 50
				burst := 10

				data := bytes.Repeat([]byte("r"), total)
				r := bytes.NewReader(data)
				th := NewThrottler(r, nil, WithReadLimit(bps, burst))

				buf := make([]byte, total)
				start := time.Now()
				n, err := th.Read(buf)
				dur := time.Since(start)

				assert.Truef(t, err == nil || err == io.EOF, "unexpected read error: %v", err)
				assert.Equal(t, total, n, "read bytes mismatch")

				expect := expectedMinDuration(total, bps, burst)
				// allow 20% timing slack for scheduler and test flakiness
				slack := expect / 5
				assert.Truef(t, dur+slack >= expect, "read duration = %v; want at least ~%v (with slack %v)", dur, expect, slack)
			},
		},
		{
			name: "WriteRateEnforced",
			run: func(t *testing.T) {
				total := 200
				bps := 50
				burst := 10

				var bufOut bytes.Buffer
				th := NewThrottler(nil, &bufOut, WithWriteLimit(bps, burst))

				data := bytes.Repeat([]byte("w"), total)
				start := time.Now()
				n, err := th.Write(data)
				dur := time.Since(start)

				assert.NoError(t, err, "unexpected write error")
				assert.Equal(t, total, n, "written bytes mismatch")

				expect := expectedMinDuration(total, bps, burst)
				slack := expect / 5
				assert.Truef(t, dur+slack >= expect, "write duration = %v; want at least ~%v (with slack %v)", dur, expect, slack)
			},
		},
		{
			name: "ConnThrottlerPassthrough",
			run: func(t *testing.T) {
				c1, c2 := net.Pipe()
				t.Cleanup(func() { c1.Close(); c2.Close() })

				// Wrap c2 with unlimited throttler
				thrConn := NewConnThrottler(c2, 0, 0, 0, 0)

				// write from c1, read from thrConn
				msg := []byte("hello-throttle")

				done := make(chan error, 1)
				go func() {
					defer c1.Close()
					_, err := c1.Write(msg)
					done <- err
				}()

				// read on wrapped conn
				got := make([]byte, len(msg))
				n, err := thrConn.Read(got)
				assert.Truef(t, err == nil || err == io.EOF, "conn read error: %v", err)
				assert.Equal(t, len(msg), n, "conn read bytes mismatch")
				assert.Equal(t, msg, got, "conn read data mismatch")

				// ensure writer had no error
				err = <-done
				assert.NoError(t, err, "writer error")
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.run(t)
		})
	}
}
