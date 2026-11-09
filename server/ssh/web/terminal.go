package web

import (
	"github.com/shellhub-io/shellhub/pkg/models"
	"golang.org/x/crypto/ssh"
)

var defaultDimensions = Dimensions{Cols: 80, Rows: 24}

type terminalMode struct {
	Key   byte
	Value uint32
}

const terminalSpeed = 14400

func (d Dimensions) ptyRequest() []byte {
	var modes []byte
	for _, mode := range []terminalMode{
		{Key: ssh.ECHO, Value: 1},
		{Key: ssh.TTY_OP_ISPEED, Value: terminalSpeed},
		{Key: ssh.TTY_OP_OSPEED, Value: terminalSpeed},
	} {
		modes = append(modes, ssh.Marshal(mode)...)
	}

	modes = append(modes, 0)

	return ssh.Marshal(models.SSHPty{
		Term:     "xterm",
		Columns:  uint32(d.Cols),
		Rows:     uint32(d.Rows),
		Width:    0,
		Height:   0,
		Modelist: modes,
	})
}
