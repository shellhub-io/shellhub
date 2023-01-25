// Package tty provides functions to write to a client TTY client.
package tty

import (
	"fmt"

	gliderssh "github.com/gliderlabs/ssh"
)

const (
	ANSIClear      = "\x1Bc"
	ANSIReset      = "\x1B[H\x1B[2J"
	ANSIBell       = "\x07c"
	ANSIBackspace  = "\x08c"
	ANSIHorizontal = "\x09c"
	ANSIVertical   = "\x0Bc"
	ANSILineFeed   = "\x0Ac"
	ANSIFormFeed   = "\x0Cc"
	ANSICarriage   = "\x0Dc"
	ANSIDelete     = "\x7Fc"
)

// Write writes the data to the client.
func Write(client gliderssh.Session, agent gliderssh.Session, data string, args ...interface{}) error {
	_, _, isPty := agent.Pty()
	if !isPty {
		return fmt.Errorf("cannot write to a non-pty client")
	}

	read, err := fmt.Fprintf(client, data, args...)
	if err != nil {
		return fmt.Errorf("failed to write to client: %w", err)
	}

	if read != len(data) {
		return fmt.Errorf("failed to write to client: %w", fmt.Errorf("failed to write all data"))
	}

	return nil
}

func Log(client gliderssh.Session, agent gliderssh.Session, msg string) error {
	return Write(client, agent, "%s\r\n", msg)
}

// Clear writes to the client the ANSI escape sequence to clear the screen.
func Clear(client gliderssh.Session, agent gliderssh.Session) error {
	return Write(client, agent, ANSIClear)
}

// Reset writes to the client the ANSI escape sequence to reset the cursor to home position (0, 0).
func Reset(client gliderssh.Session, agent gliderssh.Session) error {
	return Write(client, agent, ANSIReset)
}

// Bell writes to the client the ANSI escape sequence to ring the bell.
func Bell(client gliderssh.Session, agent gliderssh.Session) error {
	return Write(client, agent, ANSIBell)
}

// Backspace writes to the client the ANSI escape sequence to move the cursor back one space.
func Backspace(client gliderssh.Session, agent gliderssh.Session) error {
	return Write(client, agent, ANSIBackspace)
}

// Horizontal writes to the client the ANSI escape sequence to move the cursor to the next horizontal tab stop.
func Horizontal(client gliderssh.Session, agent gliderssh.Session) error {
	return Write(client, agent, ANSIHorizontal)
}

// Vertical writes to the client the ANSI escape sequence to move the cursor to the next vertical tab stop.
func Vertical(client gliderssh.Session, agent gliderssh.Session) error {
	return Write(client, agent, ANSIVertical)
}

// LineFeed writes to the client the ANSI escape sequence to move the cursor down one line.
func LineFeed(client gliderssh.Session, agent gliderssh.Session) error {
	return Write(client, agent, ANSILineFeed)
}

// FormFeed writes to the client the ANSI escape sequence to move the cursor to the next form feed.
func FormFeed(client gliderssh.Session, agent gliderssh.Session) error {
	return Write(client, agent, ANSIFormFeed)
}

// Carriage writes to the client the ANSI escape sequence to move the cursor to the beginning of the current line.
func Carriage(client gliderssh.Session, agent gliderssh.Session) error {
	return Write(client, agent, ANSICarriage)
}

// Delete writes to the client the ANSI escape sequence to delete the character at the cursor position.
func Delete(client gliderssh.Session, agent gliderssh.Session) error {
	return Write(client, agent, ANSIDelete)
}
