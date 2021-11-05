package websocket

import (
	"bytes"
	"io"
	"time"
	"unicode/utf8"

	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/websocket"
)

const pingInterval = time.Second * 30

func copyWorker(dst io.Writer, src io.Reader, doneCh chan<- bool) {
	if _, err := io.Copy(dst, src); err != nil {
		logrus.Error("could not copy the worker")
	}

	doneCh <- true
}

func redirToWs(rd io.Reader, ws *websocket.Conn) error {
	var buf [32 * 1024]byte
	var start, end, buflen int

	for {
		nr, err := rd.Read(buf[start:])
		if err != nil {
			return err
		}

		buflen = start + nr
		for end = buflen - 1; end >= 0; end-- {
			if utf8.RuneStart(buf[end]) {
				ch, width := utf8.DecodeRune(buf[end:buflen])
				if ch != utf8.RuneError {
					end += width
				}

				break
			}

			if buflen-end >= 6 {
				end = nr

				break
			}
		}

		if _, err = ws.Write([]byte(string(bytes.Runes(buf[0:end])))); err != nil {
			return err
		}

		start = buflen - end

		if start > 0 {
			// copy remaning read bytes from the end to the beginning of a buffer.
			// so that we will get normal bytes.
			for i := 0; i < start; i++ {
				buf[i] = buf[end+i]
			}
		}
	}
}

type wsconn struct {
	pinger *time.Ticker
}

func (w *wsconn) keepAlive(ws *websocket.Conn) {
	for {
		if err := ws.SetDeadline(clock.Now().Add(pingInterval * 2)); err != nil {
			return
		}

		if fw, err := ws.NewFrameWriter(websocket.PingFrame); err != nil {
			return
		} else if _, err = fw.Write([]byte{}); err != nil {
			return
		}

		if _, running := <-w.pinger.C; !running {
			return
		}
	}
}
