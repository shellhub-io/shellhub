package web

import (
	"testing"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
)

func TestDimensionsPtyRequest(t *testing.T) {
	payload := Dimensions{Cols: 132, Rows: 43}.ptyRequest()

	var pty models.SSHPty
	require.NoError(t, ssh.Unmarshal(payload, &pty))

	assert.Equal(t, "xterm", pty.Term)
	assert.Equal(t, uint32(132), pty.Columns)
	assert.Equal(t, uint32(43), pty.Rows)
	assert.Equal(t, []byte{
		53, 0, 0, 0, 1,
		128, 0, 0, 0x38, 0x40,
		129, 0, 0, 0x38, 0x40,
		0,
	}, pty.Modelist, "ECHO on, both speeds 14400, in a fixed order, so the bridge sends the bytes it parked in the handoff")
}
