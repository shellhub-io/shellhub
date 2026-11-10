// Package shellonopen is the format of a session channel's open data when the gateway asks the
// device to start the shell as it opens the channel. The device serves the requests as if they
// had followed the open, so the shell's first output leaves it one round trip earlier.
package shellonopen

import (
	"github.com/shellhub-io/shellhub/pkg/models"
	gossh "golang.org/x/crypto/ssh"
)

// Marker opens the data of a channel that starts its shell on open.
const Marker = "shell-on-open@shellhub.io"

const shellRequestType = string(models.SessionEventTypeShell)

type openData struct {
	Marker   string
	Requests []byte `ssh:"rest"`
}

type openedRequest struct {
	Type    string
	Payload []byte
	Rest    []byte `ssh:"rest"`
}

// SetsUp reports whether a request of requestType may come before the shell in the open data.
func SetsUp(requestType string) bool {
	switch models.SessionEventType(requestType) {
	case models.SessionEventTypeEnv, models.SessionEventTypePtyRequest, models.SessionEventTypeAuthAgentRequest:
		return true
	default:
		return false
	}
}

// Encode returns the open data that asks the device to serve setup and then start a shell.
func Encode(setup []*gossh.Request) []byte {
	var list []byte
	for _, req := range setup {
		list = append(list, gossh.Marshal(openedRequest{Type: req.Type, Payload: req.Payload, Rest: nil})...)
	}

	list = append(list, gossh.Marshal(openedRequest{Type: shellRequestType, Payload: nil, Rest: nil})...)

	return gossh.Marshal(openData{Marker: Marker, Requests: list})
}

// Decode returns the requests in data, ending in the shell. marked is false when data does not
// carry [Marker], and valid is false when it does but its requests are not a session setup
// ending in a shell, which the device must refuse.
func Decode(data []byte) (requests []*gossh.Request, marked, valid bool) {
	var open openData
	if err := gossh.Unmarshal(data, &open); err != nil || open.Marker != Marker {
		return nil, false, false
	}

	for list := open.Requests; len(list) > 0; {
		var req openedRequest
		if err := gossh.Unmarshal(list, &req); err != nil {
			return nil, true, false
		}

		requests = append(requests, &gossh.Request{Type: req.Type, Payload: req.Payload})
		list = req.Rest
	}

	if len(requests) == 0 || requests[len(requests)-1].Type != shellRequestType {
		return nil, true, false
	}

	for _, req := range requests[:len(requests)-1] {
		if !SetsUp(req.Type) {
			return nil, true, false
		}
	}

	return requests, true, true
}
