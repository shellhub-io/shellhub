package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"maps"
	"slices"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
)

func runOnDevice(t *testing.T, conn *ssh.Client, command string) string {
	t.Helper()

	output, err := deviceOutput(conn, command)
	require.NoError(t, err, output)

	return output
}

func deviceOutput(conn *ssh.Client, command string) (string, error) {
	sess, err := conn.NewSession()
	if err != nil {
		return "", err
	}

	defer sess.Close() //nolint:errcheck // CombinedOutput already waited for the exit status

	output, err := sess.CombinedOutput(command)

	return string(output), err
}

type scpDirectory struct {
	name  string
	mode  string
	files map[string]string
	dirs  []scpDirectory
}

func sendSCPTree(t *testing.T, conn *ssh.Client, target string, tree scpDirectory) {
	t.Helper()

	sess, err := conn.NewSession()
	require.NoError(t, err)

	defer sess.Close() //nolint:errcheck // Wait already reaped the remote scp

	stdin, err := sess.StdinPipe()
	require.NoError(t, err)

	stdout, err := sess.StdoutPipe()
	require.NoError(t, err)

	var stderr bytes.Buffer

	sess.Stderr = &stderr

	require.NoError(t, sess.Start("scp -r -t "+target))

	acks := bufio.NewReader(stdout)

	requireSCPAck(t, acks, "the sink's greeting")
	sendSCPDirectory(t, stdin, acks, tree)

	require.NoError(t, stdin.Close())
	require.NoError(t, sess.Wait(), stderr.String())
}

func sendSCPDirectory(t *testing.T, w io.Writer, acks *bufio.Reader, dir scpDirectory) {
	t.Helper()

	sendSCPMessage(t, w, acks, fmt.Sprintf("D%s 0 %s\n", dir.mode, dir.name))

	for _, name := range slices.Sorted(maps.Keys(dir.files)) {
		content := dir.files[name]

		sendSCPMessage(t, w, acks, fmt.Sprintf("C0644 %d %s\n", len(content), name))
		sendSCPMessage(t, w, acks, content+"\x00")
	}

	for _, sub := range dir.dirs {
		sendSCPDirectory(t, w, acks, sub)
	}

	sendSCPMessage(t, w, acks, "E\n")
}

func sendSCPMessage(t *testing.T, w io.Writer, acks *bufio.Reader, message string) {
	t.Helper()

	_, err := io.WriteString(w, message)
	require.NoError(t, err)

	requireSCPAck(t, acks, message)
}

func requireSCPAck(t *testing.T, acks *bufio.Reader, after string) {
	t.Helper()

	ack, err := acks.ReadByte()
	require.NoError(t, err)

	if ack != 0 {
		reason, _ := acks.ReadString('\n')
		require.Failf(t, "scp refused a message", "after %q the sink answered %d: %s", after, ack, reason)
	}
}
