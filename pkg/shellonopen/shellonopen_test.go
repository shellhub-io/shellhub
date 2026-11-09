package shellonopen

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gossh "golang.org/x/crypto/ssh"
)

func TestDecodeReadsWhatEncodeWrites(t *testing.T) {
	setup := []*gossh.Request{{Type: "env", Payload: []byte("LANG")}, {Type: "pty-req", Payload: []byte("the pty")}}

	requests, marked, valid := Decode(Encode(setup))

	assert.True(t, marked)
	assert.True(t, valid)
	require.Len(t, requests, 3)
	assert.Equal(t, "env", requests[0].Type)
	assert.Equal(t, []byte("LANG"), requests[0].Payload)
	assert.Equal(t, "pty-req", requests[1].Type)
	assert.Equal(t, []byte("the pty"), requests[1].Payload)
	assert.Equal(t, "shell", requests[2].Type)
}
