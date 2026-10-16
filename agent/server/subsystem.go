package server

import (
	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/agent/server/modes"
)

func subsystemHandlers(mode modes.Mode) map[string]gliderssh.SubsystemHandler {
	subsystemer, ok := mode.(modes.Subsystemer)
	if !ok {
		return nil
	}

	return map[string]gliderssh.SubsystemHandler{
		SFTPSubsystemName: func(session gliderssh.Session) {
			subsystemer.SFTP(session) //nolint:errcheck // a subsystem handler has no caller to report to; gliderssh ends the session when it returns
		},
	}
}
