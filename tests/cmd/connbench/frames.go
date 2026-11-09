package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

const (
	opContinuation = 0x0
	opText         = 0x1
	opBinary       = 0x2
	opClose        = 0x8
	opPing         = 0x9
	opPong         = 0xa
)

var errClosedByServer = errors.New("the server closed the websocket")

type wsChannel interface {
	receive() ([]byte, bool, error)
	sendText(data []byte) error
	Close() error
}

type frameStream struct {
	r       *bufio.Reader
	w       io.WriteCloser
	body    io.Closer
	cancel  context.CancelFunc
	writeMu sync.Mutex
}

func dialExtendedConnect(deadline time.Time, target, origin string) (*frameStream, error) {
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	reader, writer := io.Pipe()

	req, err := http.NewRequestWithContext(ctx, http.MethodConnect, target, reader)
	if err != nil {
		cancel()

		return nil, err
	}

	req.Header.Set(":protocol", "websocket")
	req.Header.Set("Sec-WebSocket-Version", "13")
	req.Header.Set("Origin", origin)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		cancel()
		_ = writer.Close()

		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		cancel()
		_ = writer.Close()
		_ = resp.Body.Close()

		return nil, fmt.Errorf("extended CONNECT answered %s", resp.Status)
	}

	return &frameStream{r: bufio.NewReader(resp.Body), w: writer, body: resp.Body, cancel: cancel}, nil
}

func (f *frameStream) receive() ([]byte, bool, error) {
	var message []byte

	binaryMessage := false

	for {
		fin, opcode, payload, err := f.readFrame()
		if err != nil {
			return nil, false, err
		}

		switch opcode {
		case opPing:
			if err := f.writeFrame(opPong, payload); err != nil {
				return nil, false, err
			}

			continue
		case opPong:
			continue
		case opClose:
			return nil, false, errClosedByServer
		case opBinary:
			binaryMessage = true
		case opText, opContinuation:
		}

		message = append(message, payload...)

		if fin {
			return message, binaryMessage, nil
		}
	}
}

func (f *frameStream) readFrame() (bool, byte, []byte, error) {
	var header [2]byte
	if _, err := io.ReadFull(f.r, header[:]); err != nil {
		return false, 0, nil, err
	}

	length := uint64(header[1] & 0x7f)

	switch length {
	case 126:
		var ext [2]byte
		if _, err := io.ReadFull(f.r, ext[:]); err != nil {
			return false, 0, nil, err
		}

		length = uint64(binary.BigEndian.Uint16(ext[:]))
	case 127:
		var ext [8]byte
		if _, err := io.ReadFull(f.r, ext[:]); err != nil {
			return false, 0, nil, err
		}

		length = binary.BigEndian.Uint64(ext[:])
	}

	payload := make([]byte, length)
	if _, err := io.ReadFull(f.r, payload); err != nil {
		return false, 0, nil, err
	}

	return header[0]&0x80 != 0, header[0] & 0x0f, payload, nil
}

func (f *frameStream) sendText(data []byte) error {
	return f.writeFrame(opText, data)
}

func (f *frameStream) writeFrame(opcode byte, payload []byte) error {
	frame := []byte{0x80 | opcode}

	switch n := len(payload); {
	case n < 126:
		frame = append(frame, 0x80|byte(n))
	case n <= 0xffff:
		frame = append(frame, 0x80|126)
		frame = binary.BigEndian.AppendUint16(frame, uint16(n))
	default:
		frame = append(frame, 0x80|127)
		frame = binary.BigEndian.AppendUint64(frame, uint64(n))
	}

	var mask [4]byte
	_, _ = rand.Read(mask[:])
	frame = append(frame, mask[:]...)

	for i, b := range payload {
		frame = append(frame, b^mask[i%4])
	}

	f.writeMu.Lock()
	defer f.writeMu.Unlock()

	_, err := f.w.Write(frame)

	return err
}

func (f *frameStream) Close() error {
	defer f.cancel()

	_ = f.w.Close()

	return f.body.Close()
}
