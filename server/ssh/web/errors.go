package web

import (
	"errors"
)

// Failures in setting up a web terminal session: resolving the device, authenticating to it,
// and opening the SSH session behind the browser's connection.
var (
	ErrPublicKey               = errors.New("failed to get the parsed public key")
	ErrConnect                 = errors.New("failed to connect to device")
	ErrSession                 = errors.New("failed to create a session between the server to the agent")
	ErrGetAuth                 = errors.New("failed to get auth data from key")
	ErrWebData                 = errors.New("failed to get the data to connect to device")
	ErrFindDevice              = errors.New("failed to find the device")
	ErrFindPublicKey           = errors.New("failed to get the public key from the server")
	ErrEvaluatePublicKey       = errors.New("failed to evaluate the public key in the server")
	ErrForbiddenPublicKey      = errors.New("failed to use the public key for this action")
	ErrDataPublicKey           = errors.New("failed to parse the public key data")
	ErrPty                     = errors.New("failed to request the pty to agent")
	ErrShell                   = errors.New("failed to get the shell to agent")
	ErrAuthentication          = errors.New("failed to authenticate to device")
	ErrInvalidVersion          = errors.New("failed to parse device version")
	ErrUnsuportedPublicKeyAuth = errors.New("connections using public keys are not permitted when the agent version is 0.5.x or earlier")
)

// Failures in framing the messages exchanged with the browser.
var (
	ErrConnReadMessageSocketWrite  = errors.New("failed to write the message's data to socket")
	ErrConnReadMessageJSONInvalid  = errors.New("failed to parse the message from json")
	ErrConnReadMessageKindInvalid  = errors.New("this kind of message is invalid")
	ErrConnWriteMessageFailedFrame = errors.New("failed to create frame")
	ErrConnReadMessageInputTooLong = errors.New("input is too long, maximum allowed is 4096 runes")
)

// Failures in reading the parameters the browser passes when upgrading to a WebSocket.
var (
	ErrWebSocketGetToken      = errors.New("failed to get the token from query")
	ErrWebSocketGetDimensions = errors.New("failed to get terminal dimensions from query")
	ErrWebSocketGetIP         = errors.New("failed to get IP from query")
)

var (
	// ErrBridgeCredentialsNotFound is returned when the SSH bridge finds no parked credentials
	// for the connection, which means the handoff expired or was already consumed.
	ErrBridgeCredentialsNotFound = errors.New("failed to find the credentials")
	// ErrBridgeReadHostKey is returned when the SSH server's host key cannot be read, leaving
	// the bridge nothing to pin its loopback connection to.
	ErrBridgeReadHostKey = errors.New("failed to read the SSH server's host key")
)

// Failures in reading the parameters of a session request.
var (
	ErrGetToken      = errors.New("token not found on request query")
	ErrGetIP         = errors.New("ip not found on request query")
	ErrGetDimensions = errors.New("failed to get a terminal dimension")
)

// ErrCreditialsNoPassword is returned when password authentication was asked for but the
// parked credentials hold none.
var ErrCreditialsNoPassword = errors.New("this creditials does not have a password defined")

// Refusals the browser is shown, as opposed to internal failures.
var (
	ErrAccessDenied = errors.New("access to the device has been denied")
	ErrInvalidSSHID = errors.New("invalid sshid format")
)
