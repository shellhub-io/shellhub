package web

import (
	"errors"
	"fmt"
)

var (
	ErrPublicKey               = fmt.Errorf("failed to get the parsed public key")
	ErrPrivateKey              = fmt.Errorf("failed to get a key data from the server")
	ErrSigner                  = fmt.Errorf("failed to create a signer from the private key")
	ErrConnect                 = fmt.Errorf("failed to connect to device")
	ErrSession                 = fmt.Errorf("failed to create a session between the server to the agent")
	ErrGetAuth                 = fmt.Errorf("failed to get auth data from key")
	ErrWebData                 = fmt.Errorf("failed to get the data to connect to device")
	ErrFindDevice              = fmt.Errorf("failed to find the device")
	ErrFindPublicKey           = fmt.Errorf("failed to get the public key from the server")
	ErrEvaluatePublicKey       = fmt.Errorf("failed to evaluate the public key in the server")
	ErrForbiddenPublicKey      = fmt.Errorf("failed to use the public key for this action")
	ErrDataPublicKey           = fmt.Errorf("failed to parse the public key data")
	ErrSignaturePublicKey      = fmt.Errorf("failed to decode the public key signature")
	ErrVerifyPublicKey         = fmt.Errorf("failed to verify the public key")
	ErrSignerPublicKey         = fmt.Errorf("failed to signer the public key")
	ErrDialSSH                 = fmt.Errorf("failed to dial to connect to server")
	ErrEnvIPAddress            = fmt.Errorf("failed to set the env variable of ip address from client")
	ErrEnvWS                   = fmt.Errorf("failed to set the env variable of web socket from client")
	ErrPipe                    = fmt.Errorf("failed to pipe client data to agent")
	ErrPty                     = fmt.Errorf("failed to request the pty to agent")
	ErrShell                   = fmt.Errorf("failed to get the shell to agent")
	ErrTarget                  = fmt.Errorf("failed to get client target")
	ErrAuthentication          = fmt.Errorf("failed to authenticate to device")
	ErrEnvs                    = fmt.Errorf("failed to parse server envs")
	ErrConfiguration           = fmt.Errorf("failed to create communication configuration")
	ErrInvalidVersion          = fmt.Errorf("failed to parse device version")
	ErrUnsuportedPublicKeyAuth = fmt.Errorf("connections using public keys are not permitted when the agent version is 0.5.x or earlier")
)

var (
	ErrConnReadMessageSocketRead  = errors.New("failed to read the message from socket")
	ErrConnReadMessageSocketWrite = errors.New("failed to write the message's data to socket")
	ErrConnReadMessageJSONInvalid = errors.New("failed to parse the message from json")
	ErrConnReadMessageKindInvalid = errors.New("this kind of message is invalid")
)

var (
	ErrWebSocketGetToken      = errors.New("failed to get the token from query")
	ErrWebSocketGetDimensions = errors.New("failed to get terminal dimensions from query")
	ErrWebSocketGetIP         = errors.New("failed to get IP from query")
)

var ErrBridgeCredentialsNotFound = errors.New("failed to find the credentials")

var (
	ErrGetToken      = errors.New("token not found on request query")
	ErrGetIP         = errors.New("ip not found on request query")
	ErrGetDimensions = errors.New("failed to get a terminal dimension")
)

var ErrCreditialsNoPassword = errors.New("this creditials does not have a password defined")
