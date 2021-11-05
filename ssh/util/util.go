package util

import (
	"io"

	"github.com/sirupsen/logrus"
)

// Write writes on a generic io.Writer and log the message if it has not been written.
func Write(io io.Writer, message string) bool {
	_, err := io.Write([]byte(message + "\r\n"))
	if err != nil {
		logrus.WithError(err).Error("could not write to the io:", message)

		return false
	}

	return true
}

// Closes a generic io.Closer and log if it has not been closed.
func Closes(io io.Closer) bool {
	if err := io.Close(); err != nil {
		logrus.WithError(err).Error("could not close the io")

		return false
	}

	return true
}

// WriteAndClose writes on a websocket.Conn, closes and log if an error occurs.
func WriteAndClose(io io.WriteCloser, message string) bool {
	if ok := Write(io, message); !ok {
		return false
	}

	if ok := Closes(io); !ok {
		return false
	}

	return true
}
